import { Emaf } from "./emaf.js";

class MerchantHeader2 extends Emaf {
    constructor(line) {
        super(line)
    }

    getMerchantName() {
        return this.get(15, 40)
    }

    getMerchantCity() {
        return this.get(40, 55)
    }

    getMerchantState() {
        return this.get(55, 58)
    }

    getMerchantZipCode() {
        return this.get(58, 67)
    }

    getMerchantCountryCode() {
        return this.get(67, 70)
    }

    getTimeZone() {
        return this.get(70, 73)
    }
}

export {
    MerchantHeader2
}