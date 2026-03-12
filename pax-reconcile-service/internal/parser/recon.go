package parser

import (
	"fmt"
	"strconv"
)

// ParseCreditReconciliation1 populates core transaction fields from a CREDIT_RECONCILIATION_1 line.
func ParseCreditReconciliation1(line string, tx *Transaction) {
	tx.DateTime = parseTxDateTime(line)
	tx.SequenceNo = field(line, 27, 36)
	tx.TypeCode = field(line, 36, 39)
	tx.AuthorizationCode = field(line, 41, 47)
	tx.EntryMode = field(line, 47, 50)
	// Last 4 digits of card account number
	acc := field(line, 50, 69)
	if len(acc) >= 4 {
		tx.AccountNumber = acc[len(acc)-4:]
	} else {
		tx.AccountNumber = acc
	}
	tx.Expiry = field(line, 69, 73)
	tx.Amount = convertAmount(field(line, 73, 84))
	tx.OldAuthorizedAmount = convertAmount(field(line, 84, 95))
	tx.CashbackAmount = convertAmount(field(line, 95, 106))
	tx.MCCSICCode = field(line, 106, 110)
	networkCode := field(line, 110, 114)
	tx.CardNetworkType = GetCardNetwork(networkCode)
	tx.CardType = GetCardNetworkType(networkCode)
	tx.ProductType = field(line, 114, 117)
	tx.DraftLocatorNo = field(line, 123, 134)
	tx.BatchNumber = field(line, 134, 140)
	tx.ConvenienceFee = convertAmount(field(line, 140, 151))
	tx.NetworkReferenceNumber = field(line, 151, 175)
	tx.TerminalNo = field(line, 175, 183)
	tx.TIC = field(line, 183, 185)
	tx.TID = field(line, 185, 188)
	tx.PinOptimization = field(line, 196, 197)
}

// parseTxDateTime converts MMDDYYYY + HHMM to an ISO-8601 UTC timestamp.
func parseTxDateTime(line string) string {
	d := field(line, 15, 23) // MMDDYYYY
	t := field(line, 23, 27) // HHMM
	if len(d) < 8 || len(t) < 4 {
		return ""
	}
	year, _ := strconv.Atoi(d[4:8])
	month, _ := strconv.Atoi(d[0:2])
	day, _ := strconv.Atoi(d[2:4])
	hour, _ := strconv.Atoi(t[0:2])
	min, _ := strconv.Atoi(t[2:4])
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:00Z", year, month, day, hour, min)
}

// ParseCreditReconciliation2 populates authorisation and network-specific fields.
func ParseCreditReconciliation2(line string, tx *Transaction) {
	tx.AuthorizationSource = field(line, 15, 16)
	tx.Indicator = field(line, 16, 17)
	tx.CAT = field(line, 17, 19)
	tx.AVSResponse = field(line, 19, 21)
	tx.CVV2 = field(line, 21, 22)
	tx.RegNo = field(line, 22, 26)
	tx.MerchantSuppliedData = field(line, 35, 44)
	tx.Currency = field(line, 44, 47)
	tx.ACI = field(line, 47, 49)
	tx.InterchangeCode = field(line, 51, 60)
	tx.InterchangeAmount = field(line, 74, 75) + convertInterchangeFee(field(line, 59, 74))
	tx.SurchargeCode = field(line, 75, 78)
	tx.SurchargeAdjustAmount = field(line, 86, 87) + convertAmount(field(line, 78, 86))
	tx.TransactionID = field(line, 87, 102)
	tx.ValidationCode = field(line, 102, 106)
	tx.AuthCode = field(line, 106, 108)
	tx.BanknetNo = field(line, 87, 96)
	tx.BanknetSettlementDate = field(line, 96, 100)
	tx.VisaProductCode = field(line, 108, 110)
	tx.RTCSettlementType = field(line, 110, 111)
	tx.Token = field(line, 111, 130)
	tx.TokenID = field(line, 130, 136)
	tx.EVMTransactionIndicator = field(line, 136, 137)
	tx.TokenExpiry = field(line, 137, 141)
	tx.AccountEndingNumber = field(line, 141, 145)
	tx.TokenAssuranceLevel = field(line, 145, 147)
	tx.AccountRefNo = field(line, 158, 193)
}

// ParseCreditReconciliation3 populates DCC/tip fields; does not overwrite existing Amount/Currency.
func ParseCreditReconciliation3(line string, tx *Transaction) {
	if tx.Amount == "" {
		tx.Amount = convertAmount(field(line, 15, 27))
	}
	if tx.Currency == "" {
		tx.Currency = field(line, 27, 30)
	}
	tx.DCCMCPIndicator = field(line, 53, 54)
	tx.TipGratuityAmount = convertAmount(field(line, 54, 65))
}

// ParseCreditReconciliation4 populates token and amount fields from alternate record.
func ParseCreditReconciliation4(line string, tx *Transaction) {
	tx.Token = field(line, 63, 82)
	tx.TokenID = field(line, 82, 88)
	if tx.Amount == "" {
		tx.Amount = convertAmount(field(line, 50, 61))
	}
	tx.ActionType = field(line, 61, 63)
}

// ParseCreditReconciliation5 extracts the Amazon Pay Charge ID.
func ParseCreditReconciliation5(line string, tx *Transaction) {
	tx.AmazonPayChargeID = field(line, 15, 42)
}

// ParseCreditReconciliation13 extracts SoftPOS mobile device fields.
func ParseCreditReconciliation13(line string, tx *Transaction) {
	tx.SoftPOSMobileDeviceType = field(line, 15, 16)
	tx.SoftPOSMobileTerminalId = field(line, 16, 48)
}

// ParseCreditReconciliation15 extracts merchant surcharge fields.
func ParseCreditReconciliation15(line string, tx *Transaction) {
	tx.MerchantSurchargeAmount = field(line, 15, 22)
	tx.MerchantSurchargeAmountSign = field(line, 22, 23)
}

// ParseCustomerID1 extracts the correlation ID.
func ParseCustomerID1(line string, tx *Transaction) {
	tx.CorrelationId = field(line, 15, 24)
}

// ParseRewardData1 extracts reward/loyalty fields.
func ParseRewardData1(line string, tx *Transaction) {
	tx.TerminalAllowsRewardsLoyalty = field(line, 15, 16)
	tx.BinEligibleForRewardsLoyalty = field(line, 16, 17)
	tx.CardEligibleForRewardsLoyalty = field(line, 17, 18)
	tx.RewardLoyaltyAmount = field(line, 18, 25)
}
