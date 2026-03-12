package parser

// ParseMerchantHeader1 extracts batch/merchant identity fields from a MID_HEADER_1 line.
func ParseMerchantHeader1(line string, mr *MerchantRecord) {
	mr.BatchSettlementType = BatchSettlementTypeMapper(field(line, 15, 17))
	mr.SettlementMID = field(line, 17, 33)
	mr.FrontEndMID = field(line, 33, 49)
	mr.DivisionNumber = field(line, 49, 52)
	mr.StoreNumber = field(line, 52, 61)
	mr.MerchantName = field(line, 61, 86)
	mr.MerchantCountryCode = field(line, 86, 88)
	mr.BatchNumberFileID = field(line, 90, 98)
}

// ParseMerchantHeader2 extracts merchant location fields from a MID_HEADER_2 line.
func ParseMerchantHeader2(line string, mr *MerchantRecord) {
	if mr.MerchantName == "" {
		mr.MerchantName = field(line, 15, 40)
	}
	mr.MerchantCity = field(line, 40, 55)
	mr.MerchantState = field(line, 55, 58)
	mr.MerchantZipCode = field(line, 58, 67)
	if mr.MerchantCountryCode == "" {
		mr.MerchantCountryCode = field(line, 67, 70)
	}
}
