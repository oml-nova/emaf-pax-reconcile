import { Emaf } from "./emaf.js";

class CreditReconciliation5 extends Emaf {
    constructor(line) {
        super(line)
    }

    getAmazonPayChargeID() {
        return this.get(15, 42)
    }

}

export {
    CreditReconciliation5
}