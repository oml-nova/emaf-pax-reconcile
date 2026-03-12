import { Emaf } from "./emaf.js";
import { convertAmount } from "../helper/helper.js";

class CreditReconciliation3 extends Emaf {
    constructor(line) {
        super(line)
    }

    getTransactionAmount() {
        const amount = this.get(15, 27)
        return convertAmount(amount)
    }

    getTransactionAmountCurrencyCode() {
        return this.get(27, 30)
    }

    getCardHolderBillingAmount() {
        return this.get(27, 30)
    }

    getCardHolderBillingAmount() {
        return this.get(27, 30)
    }

    getDCCMCPIndicator() {
        return this.get(53, 54)
    }

    getTipORGratuityAmount() {
        const amount = this.get(54, 65)
        return convertAmount(amount)
    }
}

export {
    CreditReconciliation3
}