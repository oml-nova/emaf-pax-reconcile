import { Emaf } from "./emaf.js";

class CreditReconciliation13 extends Emaf {
    constructor(line) {
        super(line);
    }

    getSoftPOSMobileDeviceType() {
        return this.get(15, 16)
    }

    getSoftPOSMobileTerminalId() {
        return this.get(16, 48)
    }
}

export {
    CreditReconciliation13
}