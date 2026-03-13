import { Emaf } from "./emaf.js";

class RewardData1 extends Emaf {
    line;
    constructor(line) {
        this.line = line;
    }

    getTerminalAllowsRewardsLoyalty() {
        return this.line.slice(15, 16)
    }

    getBinEligibleForRewardsLoyalty() {
        return this.line.slice(16, 17)
    }

    getCardEligibleForRewardsLoyalty() {
        return this.line.slice(17, 18)
    }

    getRewardLoyaltyAmount() {
        return this.line.slice(18, 25)
    }
}

export {
    RewardData1
}