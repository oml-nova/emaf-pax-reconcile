class Emaf {
    line;
    constructor(line) {
        this.line = line
    }

    get(start, end) {
        return this.line.slice(start, end)?.trim()
    }
}

export {
    Emaf
}