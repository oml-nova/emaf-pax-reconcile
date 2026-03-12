package parser

func recordTypeMapper(code string) string {
	switch code {
	case "010":
		return "CREDIT_FILE_HEADER"
	case "020":
		return "DEBIT_FILE_HEADER"
	case "030":
		return "EBT_FILE_HEADER"
	case "040":
		return "GIFTCARD_FILE_HEADER"
	case "050":
		return "POSCHECK_FILE_HEADER"
	case "060":
		return "WIC_FILE_HEADER"
	case "910":
		return "CREDIT_FILE_END"
	case "920":
		return "DEBIT_FILE_END"
	case "930":
		return "EBT_FILE_END"
	case "940":
		return "GIFTCARD_FILE_END"
	case "950":
		return "POSCHECK_FILE_END"
	case "960":
		return "WIC_FILE_END"
	case "070":
		return "MID_HEADER_1"
	case "071":
		return "MID_HEADER_2"
	case "072":
		return "MID_HEADER_3"
	case "300":
		return "CREDIT_RECONCILIATION_1"
	case "301":
		return "CREDIT_RECONCILIATION_2"
	case "302":
		return "CREDIT_RECONCILIATION_3"
	case "303":
		return "CREDIT_RECONCILIATION_4"
	default:
		return "UNKNOWN"
	}
}

func BatchSettlementTypeMapper(code string) string {
	switch code {
	case "01":
		return "CREDIT_AUTHORIZATION"
	case "02":
		return "CREDIT_DETAIL"
	case "03":
		return "CREDIT_ADJUSTMENT"
	case "21":
		return "DEBIT_AUTHORIZATION"
	case "22":
		return "DEBIT_DETAIL"
	case "23":
		return "DEBIT_ADJUSTMENT"
	case "31":
		return "EBT_AUTHORIZATION"
	case "32":
		return "EBT_DETAIL"
	case "33":
		return "EBT_ADJUSTMENT"
	case "41":
		return "GIFT_CARD_AUTHORIZATION"
	case "42":
		return "GIFT_CARD_DETAIL"
	case "51":
		return "POS_CHECK_AUTHORIZATION"
	case "52":
		return "POS_CHECK_DETAIL"
	case "61":
		return "WIC_AUTHORIZATION"
	case "62":
		return "WIC_DETAIL"
	case "63":
		return "WIC_ADJUSTMENT"
	default:
		return "UNKNOWN"
	}
}

// GetCardNetwork returns the normalised network name (e.g. "VISA", "MASTERCARD").
func GetCardNetwork(abbreviation string) string {
	switch abbreviation {
	case "DISC", "DSCV":
		return "DISCOVER"
	case "MCRD":
		return "MASTERCARD"
	case "VISA", "VIS1":
		return "VISA"
	case "AMEX":
		return "AMEX"
	default:
		return "OTHER"
	}
}

// GetCardNetworkType returns the human-readable card network name.
func GetCardNetworkType(abbreviation string) string {
	m := map[string]string{
		"AMEX": "American Express",
		"DISC": "Discover Network",
		"DSCV": "Discover Network",
		"FCOR": "FleetCor/Fuelman Auth",
		"FONE": "FleetOne Auth",
		"MCRD": "MasterCard",
		"PRVT": "Private Label",
		"PLCB": "Private Label – Cobrand",
		"VISA": "Visa",
		"VIS1": "Visa",
		"VYGR": "Voyager",
		"WEX":  "Wright Express",
		"STAR": "Star West",
		"MNY1": "Cash Station",
		"RCPA": "Star East Reciprocal",
		"AVAL": "Star East",
		"INLK": "Interlink",
		"INT2": "Interlink",
		"MAES": "Maestro",
		"CIRS": "Cirrus",
		"SHAZ": "Shazam",
		"ACCL": "Accel",
		"EBT":  "EBT Transaction",
		"GIFT": "Gift Card Transaction",
		"VSCK": "POS Check Transaction",
	}
	if name, ok := m[abbreviation]; ok {
		return name
	}
	return "Unknown"
}
