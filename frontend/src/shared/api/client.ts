import axios from 'axios'
import type { AxiosError } from 'axios'
import { ApiError } from './errors'

export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
})

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: { code: string; message: string } }>) => {
    const body = error.response?.data
    if (body?.error) {
      return Promise.reject(
        new ApiError(error.response?.status ?? 0, body.error.code, body.error.message),
      )
    }
    return Promise.reject(error)
  },
)
