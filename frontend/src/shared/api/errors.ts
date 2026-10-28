/** Normalized error thrown when the backend returns its ``{"error": {...}}`` body. */
export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, detail: string) {
    super(detail)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function getStatus(error: unknown): number | undefined {
  if (typeof error !== 'object' || error === null) return undefined
  const e = error as { status?: number; response?: { status?: number } }
  return e.status ?? e.response?.status
}