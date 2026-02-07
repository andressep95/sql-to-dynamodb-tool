import type { GetSchemasResponse, ParseSqlRequest, Conversion } from '@/models/schemas'
import { getIdToken } from './authService'

const API_BASE_URL = import.meta.env.VITE_BASE_PATH_URL || ''
const ENDPOINT_URL = import.meta.env.VITE_ENDPOINT_URL || 'prod/api/v1/schemas'

async function getHeaders(): Promise<HeadersInit> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }
  const token = await getIdToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  return headers
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (response.status === 401) {
    window.location.href = '/login'
    throw new Error('Sesión expirada. Por favor inicia sesión nuevamente.')
  }
  
  if (!response.ok) {
    const errorBody = await response.json().catch(() => null)
    const err = new Error(errorBody?.message || `Error ${response.status}`)
    ;(err as any).details = errorBody?.details || []
    ;(err as any).errorCode = errorBody?.error || 'UNKNOWN_ERROR'
    throw err
  }
  
  return response.json()
}

export const getSchemas = async (): Promise<GetSchemasResponse> => {
  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}`, {
    method: 'GET',
    headers: await getHeaders(),
  })
  return handleResponse<GetSchemasResponse>(response)
}

export const parseSql = async (request: ParseSqlRequest): Promise<Conversion> => {
  const body = {
    sqlContent: request.sqlContent,
    optimizationType: request.optimizationType ?? 'balanced',
  }

  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}`, {
    method: 'POST',
    headers: await getHeaders(),
    body: JSON.stringify(body),
  })
  return handleResponse<Conversion>(response)
}

export const getSchemaById = async (conversionId: string): Promise<Conversion> => {
  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}/${conversionId}`, {
    method: 'GET',
    headers: await getHeaders(),
  })
  return handleResponse<Conversion>(response)
}
