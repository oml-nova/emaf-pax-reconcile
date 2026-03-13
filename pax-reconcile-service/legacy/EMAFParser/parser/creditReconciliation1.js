import { Emaf } from "./emaf.js";
import { convertAmount } from "../helper/helper.js";

class CreditReconciliation1 extends Emaf {
    constructor(line) {
        super(line)
    }

    getTransactionDate() {
        return this.get(15, 23)
    }

    getTransactionTime() {
        return this.get(23, 27);
    }

    getTransactionDateTime() {
        const dateString = this.getTransactionDate();
        const timeString = this.getTransactionTime();
        return new Date(Date.UTC(dateString.slice(-4), (dateString.slice(0, 2)) - 1, dateString.slice(2, 4), timeString.slice(0, 2), timeString.slice(-2))).toISOString()
    }

    getTransactionSequenceNumber() {
        return this.get(27, 36)
    }

    getTransactionTypeCode() {
        return this.get(36, 39) //Table A4
    }

    getAuthorizationCode() {
        return this.get(41, 47)
    }

    getPOSEntryMode() {
        return this.get(47, 50) //Table A11
    }

    getCardAccountNumber() {
        return this.get(50, 69)?.slice(-4)
    }

    getExpirationDate() {
        return this.get(69, 73)
    }

    convertAmount(transcationAmount) {
        const tranactionAmountArr = transcationAmount.split('');
        tranactionAmountArr.splice(9, 0, '.');
        return Number(tranactionAmountArr.join('')).toFixed(2)
    }

    getTransactionAmount() {
        const amount = this.get(73, 84);
        return convertAmount(amount)
    }

    getOldAuthorizationAmount() {
        const amount = this.get(84, 95);
        return convertAmount(amount)
    }

    getCashBackAmount() {
        const amount = this.get(95, 106);
        return convertAmount(amount)
    }

    getMCCSICCode() {
        return this.get(106, 110)
    }

    getCardNetworkType() {
        return this.get(110, 114) // Table A-14
    }

    getCardProductType() {
        return this.get(114, 117)
    }

    // customer supplied data for mapping
    getDraftLocatorNumber() {
        return this.get(123, 134)
    }

    getBatchNumber() {
        return this.get(134, 140)
    }

    getConvenincetFee() {
        const amount = this.get(140, 151)
        return convertAmount(amount)
    }

    getNetworkReferenceNumber() {
        return this.get(151, 175)
    }

    getTerminalNumber() {
        return this.get(175, 183)
    }

    getTransactionIntegrityClassification() {
        return this.get(183, 185)
    }

    getTerminalIdentifier() {
        return this.get(185, 188)
    }

    getPinOptimisationFlag() {
        return this.get(196, 197)
    }
}

export {
    CreditReconciliation1
}