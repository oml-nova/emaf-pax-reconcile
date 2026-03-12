export function recordTypeMapper(code) {
  switch (code) {
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
      return "EBT_FILE__END"
    case "940":
      return "GIFTCARD_FILE_END"
    case "950":
      return "POSCHECK_FILE_END"
    case "960":
      return "WIC_FILE__END"
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
    case "307":
      return "CUSTOMER_ID_1"
    case "308":
      return "CUSTOMER_ID_2"
    case "309":
      return "REWARD_DATA_1"
    case "310":
      return "CREDIT_RECONCILIATION_5"
    case "312":
      return "CREDIT_RECONCILIATION_13"
    case "314":
      return "CREDIT_RECONCILIATION_15"
    default:
      return "UNKNOWN"
  }
}