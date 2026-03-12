package parser

// Transaction holds all fields parsed from EMAF credit reconciliation records.
// JSON keys match the original JS service to ensure compatibility.
type Transaction struct {
	DateTime                string `json:"date_time"`
	SequenceNo              string `json:"sequence_no"`
	TypeCode                string `json:"type_code"`
	AuthorizationCode       string `json:"authorization_code"`
	EntryMode               string `json:"entry_mode"`
	AccountNumber           string `json:"account_number"`
	Expiry                  string `json:"expiry"`
	Amount                  string `json:"amount"`
	OldAuthorizedAmount     string `json:"old_authorized_amount"`
	CashbackAmount          string `json:"cashback_amount"`
	MCCSICCode              string `json:"mcc_sic_code"`
	CardNetworkType         string `json:"card_network_type"`
	CardType                string `json:"card_type"`
	ProductType             string `json:"product_type"`
	DraftLocatorNo          string `json:"draft_locator_no"`
	BatchNumber             string `json:"batch_number"`
	ConvenienceFee          string `json:"convenience_fee"`
	NetworkReferenceNumber  string `json:"network_reference_number"`
	TerminalNo              string `json:"terminal_no"`
	TIC                     string `json:"tic"`
	TID                     string `json:"tid"`
	PinOptimization         string `json:"pin_optimization"`
	AuthorizationSource     string `json:"authorization_source"`
	Indicator               string `json:"indicator"`
	CAT                     string `json:"cat"`
	AVSResponse             string `json:"avs_response"`
	CVV2                    string `json:"cvv2"`
	RegNo                   string `json:"reg_no"`
	MerchantSuppliedData    string `json:"merchant_supplied_data"`
	Currency                string `json:"currency"`
	ACI                     string `json:"aci"`
	InterchangeCode         string `json:"interchange_code"`
	InterchangeAmount       string `json:"interchange_amount"`
	SurchargeCode           string `json:"surcharge_code"`
	SurchargeAdjustAmount   string `json:"surcharge_adjustment_amount"`
	TransactionID           string `json:"transaction_id"`
	ValidationCode          string `json:"validation_code"`
	AuthCode                string `json:"auth_code"`
	BanknetNo               string `json:"banknet_no"`
	BanknetSettlementDate   string `json:"banknet_settlement_date"`
	VisaProductCode         string `json:"visa_product_code"`
	RTCSettlementType       string `json:"rtc_settlement_type"`
	Token                   string `json:"token"`
	TokenID                 string `json:"token_id"`
	TokenExpiry             string `json:"token_expiry"`
	TokenAssuranceLevel     string `json:"token_assurance_level"`
	AccountEndingNumber     string `json:"account_ending_number"`
	EVMTransactionIndicator string `json:"evm_transaction_indicator"`
	AccountRefNo            string `json:"account_ref_no"`
	DCCMCPIndicator         string `json:"dcc_mcp"`
	TipGratuityAmount       string `json:"tip_gratuity_amount"`
	ActionType              string `json:"action_type"`
	// CR5
	AmazonPayChargeID string `json:"amazon_pay_charge_id"`
	// CR13
	SoftPOSMobileDeviceType string `json:"soft_pos_mobile_device_type"`
	SoftPOSMobileTerminalId string `json:"soft_pos_mobile_terminal_id"`
	// CR15
	MerchantSurchargeAmount     string `json:"merchant_surcharge_amount"`
	MerchantSurchargeAmountSign string `json:"merchant_surcharge_amount_sign"`
	// Customer ID
	CorrelationId string `json:"correlation_id"`
	// Reward Data
	TerminalAllowsRewardsLoyalty string `json:"terminal_allows_rewards_loyalty"`
	BinEligibleForRewardsLoyalty string `json:"bin_eligible_for_rewards_loyalty"`
	CardEligibleForRewardsLoyalty string `json:"card_eligible_for_rewards_loyalty"`
	RewardLoyaltyAmount          string `json:"reward_loyalty_amount"`
}

// MerchantRecord accumulates merchant-level context while parsing an EMAF file.
type MerchantRecord struct {
	RecordType          string
	Version             string
	ChainCode           string
	FileCreationDate    string
	BatchSettlementType string
	SettlementMID       string
	FrontEndMID         string
	DivisionNumber      string
	StoreNumber         string
	MerchantName        string
	MerchantCountryCode string
	BatchNumberFileID   string
	MerchantCity        string
	MerchantState       string
	MerchantZipCode     string
	ProcessingDate      string
	TimeZone            string
}

// SQSPayload is the message body sent downstream for reconciliation.
type SQSPayload struct {
	MerchantID  string      `json:"merchantId"`
	Transaction Transaction `json:"transaction"`
}
