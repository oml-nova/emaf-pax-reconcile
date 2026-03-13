import { Emaf } from "./emaf.js";

class CreditHeader extends Emaf {
    constructor(line) {
        super(line)
    }

    getFileVersion() {
        return this.get(23, 28)
    }

    getCreatingDate() {
        return this.get(28, 36)
    }

    getCreatingTime() {
        return this.get(36, 42)
    }

    getCreationDateTime() {
        const dateString = this.getCreatingDate();
        const timeString = this.getCreatingTime();
        return new Date(Date.UTC(dateString.slice(0, 4), dateString.slice(4, 6) - 1, dateString.slice(-2), timeString.slice(0, 2), timeString.slice(2,4), timeString.slice(-2))).toISOString()
    }

    getProcessingDate() {
        return this.get(42, 50)
    }

    getChainCode() {
        return this.get(105, 111)
    }
}

export {
    CreditHeader
}