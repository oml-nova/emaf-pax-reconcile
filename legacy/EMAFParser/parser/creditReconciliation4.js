import { Emaf } from "./emaf.js";
import { convertAmount } from "../helper/helper.js";

class CreditReconciliation4 extends Emaf {
    constructor(line) {
        super(line)
    }

    getCardAccountNumber() {
        return this.get(21, 50)
    }

    getTransactionAmount() {
        const amount = this.get(50, 61)
        return convertAmount(amount)
    }

    getActionType() {
        return this.get(61, 63)
    }

    getToken() {
        return this.get(63, 82)
    }

    getTokenId() {
        return this.get(82, 88)
    }
}

export {
    CreditReconciliation4
}