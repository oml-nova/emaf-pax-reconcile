import { Emaf } from "./emaf.js";

class CustomerID1 extends Emaf {
    constructor(line) {
        super(line)
    }

    getCorrelationId(line) {
        return this.get(15, 24)
    }
}

export {
    CustomerID1
}