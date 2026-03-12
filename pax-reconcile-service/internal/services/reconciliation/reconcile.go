package reconciliation

import (
	"bufio"
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	database "github.com/emaf-pax/reconcile-service/internal/config/databases"
	"github.com/emaf-pax/reconcile-service/internal/config/kafka"
	"github.com/emaf-pax/reconcile-service/internal/parser"
	logger "github.com/emaf-pax/reconcile-service/pkg/superlog"
)

// merchantMapping is the pre-loaded merchantId -> restaurantRefId cache.
var merchantMapping map[string]string

// LoadMerchantMappings fetches all merchant accounts for a given gateway
// and caches them in memory.
func LoadMerchantMappings(ctx context.Context) error {
	collection := os.Getenv("MERCHANT_COLLECTION")
	if collection == "" {
		collection = "merchant_accounts"
	}

	cur, err := database.GetMongoDB().Collection(collection).Find(ctx, bson.D{
		{Key: "gatewayId", Value: "WorldPay"},
	})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	mapping := make(map[string]string)
	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		merchantID, _ := doc["merchantId"].(string)
		refID, _ := doc["refId"].(string)
		if merchantID != "" {
			mapping[merchantID] = refID
		}
	}
	merchantMapping = mapping
	logger.Log().Info("Merchant mappings loaded", map[string]interface{}{
		"count": len(mapping),
	})
	return cur.Err()
}

// ProcessFile downloads an EMAF file from S3, parses it line-by-line,
// and reconciles each transaction against MongoDB.
func ProcessFile(ctx context.Context, body io.ReadCloser, fileName string) error {
	defer body.Close()

	var (
		mr           = &parser.MerchantRecord{}
		tx           = &parser.Transaction{}
		txInProgress bool
	)

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		recType := parser.RecordIdentifier(line)

		switch recType {
		case "CREDIT_FILE_HEADER":
			parser.ParseCreditHeader(line, mr)

		case "MID_HEADER_1":
			parser.ParseMerchantHeader1(line, mr)

		case "MID_HEADER_2":
			parser.ParseMerchantHeader2(line, mr)

		case "CREDIT_RECONCILIATION_1":
			// Flush the previous transaction before starting a new one
			if txInProgress {
				if err := reconcileTransaction(ctx, mr.SettlementMID, tx); err != nil {
					logger.Log().Warn("Reconcile error", map[string]interface{}{"error": err.Error()})
				}
			}
			tx = &parser.Transaction{}
			parser.ParseCreditReconciliation1(line, tx)
			txInProgress = true

		case "CREDIT_RECONCILIATION_2":
			parser.ParseCreditReconciliation2(line, tx)

		case "CREDIT_RECONCILIATION_3":
			parser.ParseCreditReconciliation3(line, tx)

		case "CREDIT_RECONCILIATION_4":
			parser.ParseCreditReconciliation4(line, tx)

		case "CREDIT_FILE_END":
			if txInProgress {
				if err := reconcileTransaction(ctx, mr.SettlementMID, tx); err != nil {
					logger.Log().Warn("Reconcile error", map[string]interface{}{"error": err.Error()})
				}
				txInProgress = false
			}
			mr = &parser.MerchantRecord{}
			tx = &parser.Transaction{}
		}
	}

	// Flush trailing transaction if file didn't end with CREDIT_FILE_END
	if txInProgress {
		if err := reconcileTransaction(ctx, mr.SettlementMID, tx); err != nil {
			logger.Log().Warn("Reconcile error (trailing)", map[string]interface{}{"error": err.Error()})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	logger.Log().Info("Completed processing file", map[string]interface{}{"file": fileName})
	return nil
}

// reconcileTransaction matches a parsed transaction against order_transactions in MongoDB.
func reconcileTransaction(ctx context.Context, merchantID string, tx *parser.Transaction) error {
	restaurantRefID := getRestaurantRefID(merchantID)

	amount, _ := strconv.ParseFloat(strings.TrimSpace(tx.Amount), 64)
	date, err := time.Parse(time.RFC3339, tx.DateTime)
	if err != nil {
		date = time.Now().UTC()
	}
	prevDay := date.AddDate(0, 0, -1)
	nextDay := date.AddDate(0, 0, 1)

	last4 := tx.AccountNumber
	cardType := tx.CardNetworkType

	filter := bson.D{
		{Key: "gatewayName", Value: "WorldPay"},
		{Key: "amount", Value: amount},
		{Key: "transactionType", Value: "Payment"},
		{Key: "cardDetails.cardNumber", Value: primitive.Regex{Pattern: last4 + "$", Options: "i"}},
		{Key: "cardDetails.cardType", Value: primitive.Regex{Pattern: cardType + "$", Options: "i"}},
		{Key: "createdDate", Value: bson.D{
			{Key: "$gte", Value: prevDay},
			{Key: "$lte", Value: nextDay},
		}},
	}

	if restaurantRefID != "" {
		filter = append(filter, bson.E{Key: "restaurantRefId", Value: restaurantRefID})
	}

	var transactionIdentifier string
	switch {
	case tx.CardNetworkType == "VISA" && tx.TransactionID != "":
		transactionIdentifier = strings.TrimSpace(tx.TransactionID)
	case tx.CardNetworkType == "MASTERCARD" && tx.BanknetNo != "":
		transactionIdentifier = strings.TrimSpace(tx.BanknetNo)
	case tx.AuthorizationCode != "":
		transactionIdentifier = strings.TrimSpace(tx.AuthorizationCode)
	}

	if transactionIdentifier != "" {
		pattern := `<TransactionIdentifier>\s*.*` + transactionIdentifier + `.*\s*<\/TransactionIdentifier>`
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "gatewayResponse", Value: primitive.Regex{Pattern: pattern, Options: "i"}}},
			bson.D{{Key: "transactionIdentifier", Value: primitive.Regex{Pattern: transactionIdentifier}}},
		}})
	}

	orderTransactions := database.GetMongoDB().Collection("order_transactions")
	var found bson.M
	err = orderTransactions.FindOne(ctx, filter).Decode(&found)

	if err == nil {
		return handleMatched(ctx, found, merchantID, restaurantRefID, transactionIdentifier, amount, date, tx)
	}
	if err == mongo.ErrNoDocuments {
		return handleUnmatched(ctx, merchantID, restaurantRefID, transactionIdentifier, amount, date, tx)
	}
	return err
}

func handleMatched(
	ctx context.Context,
	found bson.M,
	merchantID, restaurantRefID, transactionIdentifier string,
	amount float64,
	date time.Time,
	tx *parser.Transaction,
) error {
	refID, _ := found["refId"].(string)
	orderRefID, _ := found["orderRefId"].(string)
	now := time.Now().UTC()

	cardDetails := bson.M{
		"cardNumber": tx.AccountNumber,
		"cardType":   tx.CardNetworkType,
		"expiryDate": tx.Expiry,
	}

	matchedCol := database.GetMongoDB().Collection("matched_order_transaction")
	_, err := matchedCol.UpdateOne(ctx,
		bson.D{{Key: "refId", Value: refID}},
		bson.D{
			{Key: "$set", Value: bson.M{
				"refId":                 refID,
				"orderRefId":            orderRefID,
				"orderTransactionRefId": refID,
				"restaurantRefId":       restaurantRefID,
				"merchantId":            merchantID,
				"gatewayName":           "WorldPay",
				"amount":                amount,
				"transactionIdentifier": transactionIdentifier,
				"authorizationCode":     strings.TrimSpace(tx.AuthorizationCode),
				"cardDetails":           cardDetails,
				"settlementData":        tx,
				"transactionDateTime":   date,
				"lastModifiedDate":      now,
			}},
			{Key: "$setOnInsert", Value: bson.M{"createdDate": now}},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	result, err := database.GetMongoDB().Collection("order_transactions").UpdateOne(ctx,
		bson.D{{Key: "refId", Value: refID}},
		bson.D{{Key: "$set", Value: bson.M{"paymentStatus": "Success"}}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		logger.Log().Warn("order_transactions update matched 0 docs", map[string]interface{}{"refId": refID})
	}

	producerEvent := os.Getenv("PRODUCER_EVENT")
	return kafka.PublishKafkaMessage(orderRefID, producerEvent, map[string]string{
		"orderRefId":       orderRefID,
		"transactionRefId": refID,
		"paymentStatus":    "Success",
	})
}

func handleUnmatched(
	ctx context.Context,
	merchantID, restaurantRefID, transactionIdentifier string,
	amount float64,
	date time.Time,
	tx *parser.Transaction,
) error {
	now := time.Now().UTC()
	cardDetails := bson.M{
		"cardNumber": tx.AccountNumber,
		"cardType":   tx.CardNetworkType,
		"expiryDate": tx.Expiry,
	}

	unmatchedDoc := bson.M{
		"refId":                 uuid.New().String(),
		"orderRefId":            "",
		"billId":                "",
		"amount":                amount,
		"restaurantRefId":       restaurantRefID,
		"merchantId":            merchantID,
		"gatewayName":           "WorldPay",
		"transactionIdentifier": transactionIdentifier,
		"authorizationCode":     strings.TrimSpace(tx.AuthorizationCode),
		"refundAmount":          0,
		"summary":               bson.M{},
		"cardDetails":           cardDetails,
		"posSessionRefId":       "",
		"paymentDeviceId":       "",
		"createdDate":           now,
		"lastModifiedDate":      now,
		"transactionDateTime":   date,
		"settlementData":        tx,
	}

	result := database.GetMongoDB().Collection("unmatched_order_transaction").FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "amount", Value: amount},
			{Key: "transactionIdentifier", Value: transactionIdentifier},
			{Key: "transactionDateTime", Value: date},
			{Key: "cardDetails.cardNumber", Value: tx.AccountNumber},
			{Key: "cardDetails.cardType", Value: tx.CardNetworkType},
		},
		bson.D{{Key: "$setOnInsert", Value: unmatchedDoc}},
		options.FindOneAndUpdate().SetUpsert(true),
	)
	if err := result.Err(); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return nil
}

// getRestaurantRefID looks up the restaurant refId from the pre-fetched merchant mapping.
func getRestaurantRefID(merchantID string) string {
	return merchantMapping[merchantID]
}
