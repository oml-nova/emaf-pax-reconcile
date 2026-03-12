package reconciliation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	database "github.com/emaf-pax/pax-reconcile-service/internal/config/databases"
	"github.com/emaf-pax/pax-reconcile-service/internal/parser"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

// saveParsedTransaction inserts (or upserts) a parsed EMAF transaction into
// emaf_parsed_transactions. Called after a full transaction has been assembled,
// before reconciliation runs.
func saveParsedTransaction(ctx context.Context, mr *parser.MerchantRecord, tx *parser.Transaction, fileName string) (string, error) {
	now := time.Now().UTC()
	refID := uuid.New().String()

	txDate, err := time.Parse(time.RFC3339, tx.DateTime)
	if err != nil {
		txDate = now
	}

	fileCreationDate, err := time.Parse(time.RFC3339, mr.FileCreationDate)
	if err != nil {
		fileCreationDate = now
	}

	doc := bson.M{
		"refId":         refID,
		"emafFileName":  fileName,
		"createdDate":   now,
		"lastModifiedDate": now,

		"isReconciled":               false,
		"reconciledDate":             nil,
		"reconciledTransactionRefId": nil,
		"isDeleted":                  false,
		"deletedDate":                nil,

		// File / merchant header fields
		"fileRecordType":          mr.RecordType,
		"fileVersion":             mr.Version,
		"chainCode":               mr.ChainCode,
		"fileCreationDate":        fileCreationDate,
		"settlementMid":           mr.SettlementMID,
		"frontendMid":             mr.FrontEndMID,
		"merchantName":            mr.MerchantName,
		"merchantCity":            mr.MerchantCity,
		"merchantState":           mr.MerchantState,
		"merchantZipcode":         mr.MerchantZipCode,
		"merchantCountryCode":     mr.MerchantCountryCode,
		"divisionNumber":          mr.DivisionNumber,
		"storeNumber":             mr.StoreNumber,
		"batchSettlementType":     mr.BatchSettlementType,
		"batchNumberFileIdentifier": mr.BatchNumberFileID,

		// Transaction fields
		"transactionDateTime":    txDate,
		"sequenceNo":             tx.SequenceNo,
		"typeCode":               tx.TypeCode,
		"authorizationCode":      tx.AuthorizationCode,
		"entryMode":              tx.EntryMode,
		"cardLastFour":           tx.AccountNumber,
		"accountNumber":          tx.AccountNumber,
		"cardExpiry":             tx.Expiry,
		"expiry":                 tx.Expiry,
		"amount":                 tx.Amount,
		"oldAuthorizedAmount":    tx.OldAuthorizedAmount,
		"cashbackAmount":         tx.CashbackAmount,
		"convenienceFee":         tx.ConvenienceFee,
		"mccSicCode":             tx.MCCSICCode,
		"cardNetworkType":        tx.CardNetworkType,
		"cardType":               tx.CardType,
		"productType":            tx.ProductType,
		"draftLocatorNo":         tx.DraftLocatorNo,
		"batchNumber":            tx.BatchNumber,
		"networkReferenceNumber": tx.NetworkReferenceNumber,
		"terminalNo":             tx.TerminalNo,
		"tic":                    tx.TIC,
		"tid":                    tx.TID,
		"pinOptimization":        tx.PinOptimization,
		"authorizationSource":    tx.AuthorizationSource,
		"indicator":              tx.Indicator,
		"cat":                    tx.CAT,
		"avsResponse":            tx.AVSResponse,
		"cvv2":                   tx.CVV2,
		"regNo":                  tx.RegNo,
		"merchantSuppliedData":   tx.MerchantSuppliedData,
		"currency":               tx.Currency,
		"aci":                    tx.ACI,
		"interchangeCode":        tx.InterchangeCode,
		"interchangeAmount":      tx.InterchangeAmount,
		"surchargeCode":          tx.SurchargeCode,
		"surchargeAdjustmentAmount": tx.SurchargeAdjustAmount,
		"rtcSettlementType":      tx.RTCSettlementType,
		"token":                  tx.Token,
		"tokenId":                tx.TokenID,
		"tokenExpiry":            tx.TokenExpiry,
		"tokenAssuranceLevel":    tx.TokenAssuranceLevel,
		"accountEndingNumber":    tx.AccountEndingNumber,
		"evmTransactionIndicator": tx.EVMTransactionIndicator,
		"accountRefNo":           tx.AccountRefNo,
		"dccMcp":                 tx.DCCMCPIndicator,
		"tipGratuityAmount":      tx.TipGratuityAmount,
		"actionType":             tx.ActionType,
		"transactionId":          tx.TransactionID,
		"banknetNo":              tx.BanknetNo,
		"banknetSettlementDate":  tx.BanknetSettlementDate,
		"validationCode":         tx.ValidationCode,
		"visaValidationCode":     tx.ValidationCode,
		"authCode":               tx.AuthCode,
		"visaAuthCode":           tx.AuthCode,
		"visaProductCode":        tx.VisaProductCode,
	}

	col := database.GetMongoDB().Collection("emaf_parsed_transactions")
	_, err = col.InsertOne(ctx, doc)
	if err != nil {
		logger.Log().Error("Failed to save parsed EMAF transaction", map[string]interface{}{
			"error":    err.Error(),
			"refId":    refID,
			"fileName": fileName,
			"amount":   tx.Amount,
		})
		return "", err
	}
	logger.Log().Debug("Saved parsed EMAF transaction", map[string]interface{}{
		"refId":    refID,
		"fileName": fileName,
		"amount":   tx.Amount,
		"mid":      mr.SettlementMID,
	})
	return refID, nil
}

// markEmafReconciled sets isReconciled=true on the emaf_parsed_transactions doc
// identified by its refId (the value returned by saveParsedTransaction).
// reconciledTxRefId is the refId of the matched order_transaction.
//
// Required index on emaf_parsed_transactions: { refId: 1 }
func markEmafReconciled(ctx context.Context, emafRefId, reconciledTxRefId string) error {
	if emafRefId == "" {
		return nil
	}
	now := time.Now().UTC()
	col := database.GetMongoDB().Collection("emaf_parsed_transactions")
	_, err := col.UpdateOne(
		ctx,
		bson.D{{Key: "refId", Value: emafRefId}},
		bson.D{{Key: "$set", Value: bson.M{
			"isReconciled":               true,
			"reconciledDate":             now,
			"reconciledTransactionRefId": reconciledTxRefId,
			"lastModifiedDate":           now,
		}}},
	)
	if err != nil {
		logger.Log().Error("Failed to mark emaf transaction as reconciled", map[string]interface{}{
			"error": err.Error(), "emafRefId": emafRefId,
		})
		return err
	}
	return nil
}
