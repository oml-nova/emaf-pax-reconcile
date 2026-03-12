import { Emaf } from "./emaf.js";
import { convertAmount } from "../helper/helper.js";


class CreditReconciliation2 extends Emaf {
    constructor(line) {
        super(line)
    }

    getAuthorizationSource() {
        return this.get(15, 16) // Table A-1
    }

    getMailPhoneIndicator() {
        return this.get(16, 17) // Table A-8
    }

    getCATIndicator() {
        return this.get(17, 19) // Table A-10
    }

    getASVResponseCode() {
        return this.get(19, 21) // Table A-5
    }

    getCVV2ResponseIndicator() {
        return this.get(21, 22) // Table A-7
    }

    getRegistrationNumber() {
        return this.get(22, 26)
    }

    getTerminalNumber() {
        return this.get(26, 35)
    }

    getMerchatSuppliedData() {
        return this.get(35, 44)
    }

    getCurrencyCode() {
        return this.get(44, 47) // Table A-15
    }

    getAuthorizationCharacteristicIndicator() {
        return this.get(47, 49) // Table A-9
    }

    getOriginalInterchangeIndicator() {
        return this.get(49, 51) // Table A-12
    }

    getInterChangeCode() {
        return this.get(51, 60) // Table A-128
    }

    getInterChangeReimbursenebtFee() {
        const amount = this.get(59, 74);
        const tranactionAmountArr = amount.split('');
        tranactionAmountArr.splice(5, 0, '.');
        return Number(tranactionAmountArr.join('')).toFixed(9)
    }

    getInterChangeSign() {
        return this.get(74, 75)
    }

    getSurChargeReasonOrInterchangeAdjustmentReasonCode() {
        return this.get(75, 78) // Table A-13
    }

    getSurChargeOrInterchangeAdjustmentAmount() {
        const amount = this.get(78, 86);
        return convertAmount(amount)
    }

    getSurchargeInterChangeAdjustmentSign() {
        return this.get(86, 87)
    }

    getSurchargeInterChangeAdjustmentSign() {
        return this.get(86, 87)
    }

    getVisaTransactionId() {
        return this.get(87, 102)
    }

    getVisaValidationCode() {
        return this.get(102, 106)
    }

    getVisaAuthorizationCode() {
        return this.get(106, 108)
    }

    getMasterCardBanknetRefNumber() {
        return this.get(87, 96)
    }

    getMasterCardBanknetSettlementDate() {
        return this.get(96, 100)
    }

    getMasterCardReserved() {
        return this.get(100, 108)
    }

    getVisaProductCode() {
        return this.get(108, 110) // Table A-31
    }

    getRTCSettlementType() {
        return this.get(110, 111)
    }

    getToken() {
        return this.get(111, 130)
    }

    getTokenId() {
        return this.get(130, 136)
    }

    getEVMTransactionIndicator() {
        return this.get(136, 137)
    }

    getTokenExpirationDate() {
        return this.get(137, 141)
    }

    getPANEndingNumber() {
        return this.get(141, 145)
    }

    getTokenAssuranceLevel() {
        return this.get(145, 147)
    }

    getPaymentAccountReferenceNumber() {
        return this.get(158, 193)
    }
}

export {
    CreditReconciliation2
}