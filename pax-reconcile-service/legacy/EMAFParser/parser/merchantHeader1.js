import { batchSettlementTypeMapper } from "./batchSettlementTypeMapper.js";
import { Emaf } from "./emaf.js";

class MerchantHeader1 extends Emaf {
    constructor(line) {
        super(line)
    }

    getBatchSettlementType() {
        return batchSettlementTypeMapper(this.get(15, 17))
    }

    getSettlementMID() {
        return this.get(17, 33)
    }

    getFrontEndMID() {
        return this.get(33, 49)
    }

    getMerchantDivisionNumber() {
        return this.get(49, 52)
    }

    getMerchantStoreNumber() {
        return this.get(52, 61)
    }

    getMerchantName() {
        return this.get(61, 86)
    }

    getMerchantLocationCountryCode() {
        return this.get(86, 88)
    }

    getBatchNumberFileIdentifier() {
        return this.get(90, 98)
    }
}

export {
    MerchantHeader1
}