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
	case "305":
		return "CREDIT_RECONCILIATION_5"
	case "307":
		return "CUSTOMER_ID_1"
	case "308":
		return "CUSTOMER_ID_2"
	case "309":
		return "REWARD_DATA_1"
	case "312":
		return "CREDIT_RECONCILIATION_13"
	case "314":
		return "CREDIT_RECONCILIATION_15"
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
	case "DISC", "DISCV":
		return "DISCOVER"
	case "MCRD":
		return "MASTERCARD"
	case "VISA":
		return "VISA"
	case "AMEX":
		return "AMEX"
	default:
		return "OTHER"
	}
}

// GetCardNetworkType returns the human-readable card network name.
// All mappings match the legacy JS cardTypeMapper.getCardNetworkType().
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
		// Star Networks
		"STAR": "Star West",
		"MNY1": "Cash Station",
		"RCPA": "Star East Reciprocal",
		"AVAL": "Star East",
		"MAC2": "Star Northeast",
		"MAC3": "Star Northeast",
		"EXPL": "Explore/Star",
		"BKMT": "Star Midwest",
		"MACG": "MAC/Gr Machine",
		"MMAC": "MAC",
		"HNMN": "Star East",
		"CIXT": "Cash St Issuer",
		"MC3D": "MAC Conversion",
		"EXSS": "Explore Subswitch",
		"SESS": "Star East Subswitch",
		"SWSS": "Star West Subswitch",
		"EXP3": "Star West",
		"EXP4": "Star West",
		"STGY": "Star West Gateway",
		"SN1G": "Star NE POS 1 Groc",
		"SN1N": "Star NE POS 1 Non-Groc",
		"SN2G": "Star NE POS 2 Groc",
		"SN2N": "Star NE POS 2 Non-Groc",
		"SN3G": "Star NE POS 3 Groc",
		"SN3N": "Star NE POS 3 Non-Groc",
		"SS1G": "Star W POS 1 Groc",
		"SS1N": "Star SE POS 1 Non-Groc",
		"SS2G": "Star SE POS 2 Groc",
		"SS2N": "Star SE POS 2 Non-Groc",
		"SS3G": "Star SE POS 3 Groc",
		"SS3N": "Star SE POS 3 Non-Groc",
		"SW1G": "Star W POS 1 Groc",
		"SW1N": "Star W POS 1 Non-Groc",
		"SW2G": "Star W POS 2 Groc",
		"SW2N": "Star W POS 2 Non-Groc",
		"SW3G": "Star W POS 3 Groc",
		"SW3N": "Star W POS 3 Non-Groc",
		"MAC1": "MAC",
		"SEST": "Star East",
		"EXP6": "Star West",
		"EXP5": "Star West",
		"SN1P": "Star NE POS 1 Petro",
		"SN2P": "Star NE POS 2 Petro",
		"SN3P": "Star NE POS 3 Petro",
		"SS1P": "Star SE POS 1 Petro",
		"SS2P": "Star SE POS 2 Petro",
		"SS3P": "Star SE POS 3 Petro",
		"SW1P": "Star W POS 1 Petro",
		"SW2P": "Star W POS 2 Petro",
		"SW3P": "Star W POS 3 Petro",
		"SN1B": "Star NE POS 1 AP BL Pay",
		"SN2B": "Star NE POS 2 AP BL Pay",
		"SN3B": "Star NE POS 3 AP BL Pay",
		"SS1B": "Star SE POS 1 AP BL Pay",
		"SS2B": "Star SE POS 2 AP BL Pay",
		"SS3B": "Star SE POS 3 AP BL Pay",
		"SW1B": "Star W POS 1 AP BL Pay",
		"SW2B": "Star W POS 2 AP BL Pay",
		"SW3B": "Star W POS 3 AP BL Pay",
		"SNSB": "Star NE BP Stnd",
		"SSSB": "Star SE BP Stnd",
		"SWSB": "Star W BP Stnd",
		"SNEB": "Star NE BP Emerging",
		"SSEB": "Star SE BP Emerging",
		"SWEB": "Star W BP Emerging",
		"SNST": "Star NE POS Small Ticket",
		"SSST": "Star SE POS Small Ticket",
		"SWST": "Star W POS Small Ticket",
		"SNMR": "Star NE POS Medical",
		"SSMR": "Star SE POS Medical",
		"SWMR": "Star W POS Medical",
		// Interlink
		"INLK": "Interlink",
		"INT2": "Interlink",
		"INS1": "Interlink Stnd SMF1",
		"INS2": "Interlink Stnd SMF2",
		"INS3": "Interlink Stnd SMF3",
		"ING1": "Interlink Groc SMF1",
		"ING2": "Interlink Groc SMF2",
		"ING3": "Interlink Groc SMF3",
		// Pulse
		"PUL2": "Pulse",
		"TYME": "TYME",
		"TYM2": "TYME 2",
		"MS1":  "Pulse -Money Station",
		"PULP": "Pulse",
		"MSZA": "Pulse Zone A",
		"MSZB": "Pulse Zone B",
		"MONY": "Money Station USB",
		"MSRN": "Pulse Reciprocal NYCE",
		"MSRS": "Pulse Reciprocal Star",
		"MSRT": "Pulse Reciprocal TYME",
		"MS9A": "Pulse Zone A Intsw",
		"MS9B": "Pulse Zone B Intsw",
		"PUZA": "Pulse Gateway Zone A",
		"PUZB": "Pulse Gateway Zone B",
		"PURN": "Pulse Gtwy Rcip NYCE",
		"PURS": "Pulse Gtwy Rcip Star",
		"PURT": "Pulse Gtwy Rcip TYME",
		"PU9A": "Pulse Gtwy A Intsw",
		"PU9B": "Pulse Gtwy B Intsw",
		"PNSR": "Pulse Limited Recip",
		"MNSR": "Pulse-MSI Ltd Recip",
		"PULS": "NCR Pulse",
		"PUGU": "Pulse-Gulfnet Reciprocal",
		"PDSD": "Pulse Discover Check Card",
		// NYCE
		"PSI2": "Plus Network",
		"NYC1": "NYCE East",
		"MGCL": "NYCE Midwest",
		"NYGY": "NYCE Gateway",
		// Jeanie
		"JENI": "Jeanie",
		"JPIX": "Exempt Jeanie POS",
		"JEPR": "Jeanie Presto",
		"JENA": "Jeanie Audio Network",
		"MJI1": "MPS Jeanie Int'l",
		"MJN1": "Jeanie RIC",
		"JEST": "Jeanie Northeast",
		"JNIT": "Jeanie Internet",
		"JNPB": "Jeanie Point of Bank",
		"JELM": "Jeanie LM",
		"JADV": "Jeanie Advtg",
		"JNIB": "Jeanie",
		"JPRB": "Jeanie Presto",
		"JPXB": "Exmpt Jeanie POS",
		"JLMB": "Jeanie LM",
		"JPBB": "Jeanie POB",
		"MJN2": "Jeanie RIC",
		"JADB": "Jeanie Advtg",
		"JP53": "Jeanie Preferred",
		"JGT1": "Jeni T1 GR",
		"JGT2": "Jeni T2 GR",
		"JGT3": "Jeni T3 GR",
		"JOT1": "Jeni T1 OT",
		"JOT2": "Jeni T2 OT",
		"JOT3": "Jeni T3 OT",
		"JQSR": "JENI QSR",
		"JMED": "Jeni Med",
		"AJN1": "AFFN/ Jeanie 1",
		"AJN2": "AFFN/ Jeanie 2",
		"JAF1": "Jeanie /AFFN 1",
		"JAF2": "Jeanie / AFFN 2",
		// Other Networks
		"REVM": "Revolution Money",
		"CIRS": "Cirrus",
		"SHAZ": "Shazam",
		"MAES": "Maestro",
		"ACCL": "Accel",
		"AFN1": "Armed Forces Network",
		"AFN2": "AFFN- 5/2 Switch",
		"AFMM": "AFFN MM",
		"AFDM": "AFFN DT MM",
		"AFPM": "AFFN Preferred Merchant",
		"AFFR": "AFFN Reversals",
		"AF2R": "AFN2 Reversals",
		"AML1": "AFFN/Maestro",
		"AFDP": "AFFN DT PM",
		"AML2": "AFFN/Maestro RCPRCLK",
		"APR1": "AFN1 Presto!",
		"APR2": "AFN2 Presto!",
		"AKOP": "Alaska Option",
		"CTF1": "CU 24",
		"GDEB": "Generic Debit",
		"ITEL": "Instant Teller",
		"KETS": "Kansas Elect Trans",
		"MPCT": "MPACT",
		"TX00": "TX",
		"DBT1": "Debitman",
		"PINP": "Pin Prompting",
		"PST1": "Presto!",
		// Maestro Tiers
		"MST1": "MAES Spmkt Tier 1",
		"MST2": "MAES Spmkt Tier 2",
		"MSBA": "MAES Spmkt Base",
		"MCT1": "MAES Conv Tier 1",
		"MCT2": "MAES Conv Tier 2",
		"MCBA": "MAES Conv Base",
		"MOT1": "MAES All Other Tier 1",
		"MOT2": "MAES All Other Tier 2",
		"MOBA": "MAES All Other Base",
		// Check
		"ECP":  "Bankserv",
		"ECPG": "Total Check -Guarantee",
		"ECPT": "Total Check -Truncated",
		"ECPV": "Total Check Verification",
		"BDSD": "Batch Discover",
		// EBT
		"EBT":  "EBT Transaction",
		"ENJ1": "New Jersey EBT",
		"ETX1": "Texas EBT",
		"EIL1": "Illinois EBT",
		"EOK1": "Oklahoma EBT",
		"EWCS": "AA Western Coalition EBT",
		"ENCS": "AA NE Coalition EBT",
		"ESAS": "AA Southern Alliance EBT",
		"ELA1": "Louisiana EBT",
		"EMD1": "Maryland EBT",
		"ESC1": "South Carolina EBT",
		"ENM1": "New Mexico EBT",
		"ETN1": "Tennessee EBT",
		"EUT1": "Utah EBT",
		"EQST": "AA-Quest Coalition EBT",
		"EDC1": "District of Columbia EBT",
		"EMI1": "Michigan EBT",
		"EIN1": "Indiana EBT",
		"EOR1": "Oregon EBT",
		"EDP1": "CA -San Diego EBT",
		"EDO1": "CA-San Bernardino EBT",
		"EVT1": "Vermont EBT",
		"ERI1": "Rhode Island EBT",
		"ENC1": "North Carolina EBT",
		"ENH1": "New Hampshire EBT",
		"EMN1": "Minnesota EBT",
		"EPA1": "Pennsylvania EBT",
		"EKS1": "Kansas EBT",
		"EWI1": "Wisconsin EBT",
		"ENS1": "North/South Dakota EBT",
		"EHI1": "Hawaii EBT",
		"EID1": "Idaho EBT",
		"EAK1": "Alaska EBT",
		"EWA1": "Washington EBT",
		"EAZ1": "Arizona EBT",
		"ECO1": "Colorado EBT",
		"EMA1": "Massachusetts EBT",
		"ECT1": "Connecticut EBT",
		"EME1": "Maine EBT",
		"ENY1": "New York EBT",
		"EAL1": "Alabama EBT",
		"EAR1": "Arkansas EBT",
		"EFL1": "Florida EBT",
		"EGA1": "Georgia EBT",
		"EKY1": "Kentucky EBT",
		"EMO1": "Missouri EBT",
		"EMS1": "Mississippi EBT",
		"EMT1": "Montana EBT",
		"ECA1": "California EBT",
		"EPR1": "Puerto Rico EBT",
		"EPR2": "Puerto Rico EBT",
		"EPR3": "Puerto Rico EBT",
		"ENE1": "Nebraska EBT",
		"EVA1": "Virginia EBT",
		"EDE1": "Delaware EBT",
		"ENV1": "Nevada EBT",
		"EWV1": "West Virginia EBT",
		"EIA1": "Iowa EBT",
		"SZIA": "EBT Iowa Shazam",
		"EVI1": "Virgin Islands EBT",
		"EALL": "EBT All States",
		"EGU1": "Guam EBT",
		"PRFS": "EBT Puerto Rico - FS",
		"PRCS": "EBT Puerto Rico - CS",
		"EOH1": "Ohio EBT",
		"EWY1": "EBT Wyoming",
		// Gift/POS Check/Other
		"GIFT": "Gift Card Transaction",
		"VSCK": "POS Check Transaction",
		"EHH1": "ECHO",
		"VPLN": "Visa ReadyLink",
		"UNKN": "Unknown",
		// Bill-Me-Later
		"BMLA": "Bill-Me-Later Core",
		"BMLB": "Bill-Me-Later 90 Days Same As Cash (SAC)",
		"BMLC": "Bill-Me-Later Business Core",
		"BMLD": "Bill-Me-Later Business 90 Days Same As Cash (SAC)",
		"BMLE": "Bill-Me-Later Private Label Core",
		"BMLF": "Bill-Me-Later Private Label 90 Days Same As Cash (SAC)",
	}
	if name, ok := m[abbreviation]; ok {
		return name
	}
	return "Unknown"
}
