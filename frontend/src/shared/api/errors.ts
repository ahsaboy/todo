export class ApiError extends Error {
  public readonly code: string
  public readonly status: number

  constructor(params: { message: string; code: string; status: number }) {
    super(params.message)
    this.name = 'ApiError'
    this.code = params.code
    this.status = params.status

    Object.setPrototypeOf(this, ApiError.prototype)
  }
}
