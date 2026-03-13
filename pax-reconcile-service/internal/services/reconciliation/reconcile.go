package reconciliation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	database "github.com/emaf-pax/pax-reconcile-service/internal/config/databases"
	"github.com/emaf-pax/pax-reconcile-service/internal/config/kafka"
	"github.com/emaf-pax/pax-reconcile-service/internal/parser"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

// merchantCache provides thread-safe lazy caching of merchantId -> restaurantRefId.
var merchantCache sync.Map

// ReconcileSummary holds the outcome counts for a single file run.
type ReconcileSummary struct {
	File      string `json:"file"`
	Total     int    `json:"total"`
	Matched   int    `json:"matched"`
	Unmatched int    `json:"unmatched"`
	Errors    int    `json:"errors"`
	Duration  string `json:"duration"`
}

// ProcessFile downloads an EMAF file from S3, parses it line-by-line,
// and reconciles each transaction against MongoDB.
func ProcessFile(ctx context.Context, body io.ReadCloser, fileName string) (*ReconcileSummary, error) {
	defer body.Close()
	startTime := time.Now()

	logger.Log().Info("Starting file processing", map[string]interface{}{"file": fileName})

	var (
		mr           = &parser.MerchantRecord{}
		tx           = &parser.Transaction{}
		txInProgress bool
		totalTx      int
		matchedTx    int
		unmatchedTx  int
		errorTx      int
	)

	reconcileTx := func(merchantID string) {
		totalTx++
		emafRefId, err := saveParsedTransaction(ctx, mr, tx, fileName)
		if err != nil {
			errorTx++
			logger.Log().Error("Failed to save parsed transaction, skipping reconciliation", map[string]interface{}{
				"error":      err.Error(),
				"file":       fileName,
				"txIndex":    totalTx,
				"merchantId": merchantID,
				"amount":     tx.Amount,
			})
			return
		}

		logger.Log().Debug("Reconciling transaction", map[string]interface{}{
			"file":       fileName,
			"txIndex":    totalTx,
			"merchantId": merchantID,
			"amount":     tx.Amount,
			"cardLast4":  tx.AccountNumber,
			"cardType":   tx.CardNetworkType,
			"authCode":   tx.AuthorizationCode,
			"dateTime":   tx.DateTime,
		})

		matched, err := reconcileTransaction(ctx, merchantID, tx, emafRefId)
		if err != nil {
			errorTx++
			logger.Log().Error("Reconcile failed", map[string]interface{}{
				"error":      err.Error(),
				"file":       fileName,
				"txIndex":    totalTx,
				"merchantId": merchantID,
				"amount":     tx.Amount,
				"cardLast4":  tx.AccountNumber,
			})
		} else if matched {
			matchedTx++
		} else {
			unmatchedTx++
		}
	}

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		recType := parser.RecordIdentifier(line)

		switch recType {
		case "CREDIT_FILE_HEADER":
			parser.ParseCreditHeader(line, mr)

		case "MID_HEADER_1":
			parser.ParseMerchantHeader1(line, mr)
			logger.Log().Debug("Parsed merchant header", map[string]interface{}{
				"file":          fileName,
				"settlementMID": mr.SettlementMID,
				"merchantName":  mr.MerchantName,
			})

		case "MID_HEADER_2":
			parser.ParseMerchantHeader2(line, mr)

		case "CREDIT_RECONCILIATION_1":
			if txInProgress {
				reconcileTx(mr.SettlementMID)
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

		case "CREDIT_RECONCILIATION_5":
			parser.ParseCreditReconciliation5(line, tx)

		case "CUSTOMER_ID_1":
			parser.ParseCustomerID1(line, tx)

		case "REWARD_DATA_1":
			parser.ParseRewardData1(line, tx)

		case "CREDIT_RECONCILIATION_13":
			parser.ParseCreditReconciliation13(line, tx)

		case "CREDIT_RECONCILIATION_15":
			parser.ParseCreditReconciliation15(line, tx)

		case "CREDIT_FILE_END":
			if txInProgress {
				reconcileTx(mr.SettlementMID)
				txInProgress = false
			}
			mr = &parser.MerchantRecord{}
			tx = &parser.Transaction{}
		}
	}

	// Flush trailing transaction if file didn't end with CREDIT_FILE_END
	if txInProgress {
		reconcileTx(mr.SettlementMID)
	}

	if err := scanner.Err(); err != nil {
		logger.Log().Error("File scan error", map[string]interface{}{"file": fileName, "error": err.Error()})
		return nil, err
	}

	summary := &ReconcileSummary{
		File:      fileName,
		Total:     totalTx,
		Matched:   matchedTx,
		Unmatched: unmatchedTx,
		Errors:    errorTx,
		Duration:  time.Since(startTime).String(),
	}

	logger.Log().Info("Completed processing file", map[string]interface{}{
		"file":      fileName,
		"total":     totalTx,
		"matched":   matchedTx,
		"unmatched": unmatchedTx,
		"errors":    errorTx,
		"duration":  summary.Duration,
	})
	return summary, nil
}

// reconcileTransaction matches a parsed transaction against order_transactions in MongoDB.
// Returns (true, nil) for matched, (false, nil) for unmatched, (false, err) on error.
func reconcileTransaction(ctx context.Context, merchantID string, tx *parser.Transaction, emafRefId string) (bool, error) {
	restaurantRefID := getRestaurantRefID(ctx, merchantID)

	amount, err := strconv.ParseFloat(strings.TrimSpace(tx.Amount), 64)
	if err != nil {
		return false, fmt.Errorf("invalid transaction amount %q: %w", tx.Amount, err)
	}
	date, err := time.Parse(time.RFC3339, tx.DateTime)
	if err != nil {
		logger.Log().Warn("Invalid transaction date, using current time", map[string]interface{}{
			"dateTime":   tx.DateTime,
			"merchantId": merchantID,
		})
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
		if err := handleMatched(ctx, found, merchantID, restaurantRefID, transactionIdentifier, amount, date, tx, emafRefId); err != nil {
			return false, err
		}
		return true, nil
	}
	if err == mongo.ErrNoDocuments {
		logger.Log().Info("Transaction unmatched", map[string]interface{}{
			"merchantId":            merchantID,
			"amount":                amount,
			"cardLast4":             last4,
			"cardType":              cardType,
			"transactionIdentifier": transactionIdentifier,
			"restaurantRefId":       restaurantRefID,
		})
		if err := handleUnmatched(ctx, merchantID, restaurantRefID, transactionIdentifier, amount, date, tx); err != nil {
			return false, err
		}
		return false, nil
	}
	return false, fmt.Errorf("MongoDB query error: %w", err)
}

func handleMatched(
	ctx context.Context,
	found bson.M,
	merchantID, restaurantRefID, transactionIdentifier string,
	amount float64,
	date time.Time,
	tx *parser.Transaction,
	emafRefId string,
) error {
	orderTransactionRefId, _ := found["refId"].(string)
	orderRefID, _ := found["orderRefId"].(string)
	now := time.Now().UTC()

	logger.Log().Info("Transaction matched", map[string]interface{}{
		"orderTransactionRefId": orderTransactionRefId,
		"orderRefId":            orderRefID,
		"merchantId":            merchantID,
		"amount":                amount,
		"transactionIdentifier": transactionIdentifier,
		"restaurantRefId":       restaurantRefID,
	})

	cardDetails := bson.M{
		"cardNumber": tx.AccountNumber,
		"cardType":   tx.CardNetworkType,
		"expiryDate": tx.Expiry,
	}

	matchedRefId := uuid.New().String()
	matchedCol := database.GetMongoDB().Collection("matched_order_transaction")
	_, err := matchedCol.UpdateOne(ctx,
		bson.D{{Key: "orderTransactionRefId", Value: orderTransactionRefId}},
		bson.D{
			{Key: "$set", Value: bson.M{
				"orderRefId":            orderRefID,
				"orderTransactionRefId": orderTransactionRefId,
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
			{Key: "$setOnInsert", Value: bson.M{
				"refId":       matchedRefId,
				"createdDate": now,
			}},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logger.Log().Error("Failed to upsert matched_order_transaction", map[string]interface{}{
			"error":                 err.Error(),
			"orderTransactionRefId": orderTransactionRefId,
		})
		return err
	}

	result, err := database.GetMongoDB().Collection("order_transactions").UpdateOne(ctx,
		bson.D{{Key: "refId", Value: orderTransactionRefId}},
		bson.D{{Key: "$set", Value: bson.M{"paymentStatus": "Success"}}},
	)
	if err != nil {
		logger.Log().Error("Failed to update order_transactions paymentStatus", map[string]interface{}{
			"error": err.Error(),
			"refId": orderTransactionRefId,
		})
		return err
	}
	if result.MatchedCount == 0 {
		logger.Log().Warn("order_transactions update matched 0 docs — possible stale read", map[string]interface{}{"refId": orderTransactionRefId})
	}

	producerEvent := os.Getenv("PRODUCER_EVENT")
	if err := kafka.PublishKafkaMessage(orderRefID, producerEvent, map[string]string{
		"orderRefId":       orderRefID,
		"transactionRefId": orderTransactionRefId,
		"paymentStatus":    "Success",
	}); err != nil {
		logger.Log().Error("Failed to publish Kafka reconciliation event", map[string]interface{}{
			"error":      err.Error(),
			"orderRefId": orderRefID,
			"refId":      orderTransactionRefId,
		})
		return err
	}

	logger.Log().Debug("Kafka reconciliation event published", map[string]interface{}{
		"orderRefId": orderRefID,
		"refId":      orderTransactionRefId,
		"event":      producerEvent,
	})

	if err := markEmafReconciled(ctx, emafRefId, matchedRefId); err != nil {
		return err
	}
	return nil
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
		logger.Log().Error("Failed to upsert unmatched_order_transaction", map[string]interface{}{
			"error":                 err.Error(),
			"merchantId":            merchantID,
			"transactionIdentifier": transactionIdentifier,
		})
		return err
	}
	return nil
}

// getRestaurantRefID fetches the restaurant refId for a merchantId, caching the result.
func getRestaurantRefID(ctx context.Context, merchantID string) string {
	if val, ok := merchantCache.Load(merchantID); ok {
		logger.Log().Debug("Merchant cache hit", map[string]interface{}{
			"merchantId": merchantID,
			"refId":      val.(string),
		})
		return val.(string)
	}

	var doc bson.M
	err := database.GetMongoDB().Collection("merchant_accounts").FindOne(ctx, bson.D{
		{Key: "gatewayId", Value: "WorldPay"},
		{Key: "merchantId", Value: merchantID},
	}).Decode(&doc)

	var refID string
	if err == nil {
		refID, _ = doc["refId"].(string)
		logger.Log().Info("Merchant mapping fetched from DB", map[string]interface{}{
			"merchantId": merchantID,
			"refId":      refID,
		})
	} else {
		logger.Log().Warn("Merchant mapping not found", map[string]interface{}{
			"merchantId": merchantID,
			"error":      err.Error(),
		})
	}

	merchantCache.Store(merchantID, refID)
	return refID
}
