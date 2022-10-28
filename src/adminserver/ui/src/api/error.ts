export class ApiError extends Error {
    status: number

    constructor(code: number, msg: string) {
        super(msg);
        this.status = code
        // Set the prototype explicitly.
        Object.setPrototypeOf(this, ApiError.prototype);
    }

    Message() {
        return this.message
    }

    Code() {
        return this.status
    }
}