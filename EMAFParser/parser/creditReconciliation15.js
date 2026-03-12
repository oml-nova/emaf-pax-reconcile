import { Emaf } from "./emaf.js";

class CreditReconciliation15 extends Emaf {
    constructor(line) {
        super(line)
    }

    getMerchantSurchargeAmount() {
        return this.get(15, 22)
    }

    getMerchantSurchargeAmountSign() {
        return this.get(22, 23)
    }
}

export {
    CreditReconciliation15
}