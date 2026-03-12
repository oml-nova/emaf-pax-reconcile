function batchSettlementTypeMapper(code) {
  switch (code) {
    case "01":
      return "CREDIT_AUTHORIZATION";
    case "02":
      return "CREDIT_DETAIL";
    case "03":
      return "CREDIT_ADJUSTMENT";
    case "21":
      return "DEBIT_AUTHORIZATION";
    case "22":
      return "DEBIT_DETAIL";
    case "23":
      return "DEBIT_ADJUSTMENT";
    case "31":
      return "EBT_AUTHORIZATION";
    case "32":
      return "EBT_DETAIL";
    case "33":
      return "EBT_ADJUSTMENT";
    case "41":
      return "GIFT_CARD_AUTHORIZATION";
    case "42":
      return "GIFT_CARD_DETAIL";
    case "51":
      return "POS_CHECK_AUTHORIZATION";
    case "52":
      return "POS_CHECK_DETAIL";
    case "61":
      return "WIC_AUTHORIZATION";
    case "62":
      return "WIC_DETAIL";
    case "63":
      return "WIC_ADJUSTMENT";
    default:
      return "UNKNOWN";
  }
}

export {
  batchSettlementTypeMapper
}