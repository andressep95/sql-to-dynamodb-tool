import type { GetSchemasResponse, ParseSqlRequest, Conversion } from '@/models/schemas'

const API_BASE_URL = import.meta.env.VITE_BASE_PATH_URL
const ENDPOINT_URL = import.meta.env.VITE_ENDPOINT_URL

export const getSchemas = async (): Promise<GetSchemasResponse> => {
  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  if (!response.ok) {
    throw new Error('Failed to fetch schemas')
  }
  return response.json()
}

export const parseSql = async (request: ParseSqlRequest): Promise<Conversion> => {
  const body = {
    sqlContent: request.sqlContent,
    optimizationType: request.optimizationType ?? 'balanced',
  }

  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
  if (!response.ok) {
    throw new Error('Failed to parse SQL')
  }
  return response.json()
}

export const getSchemaById = async (conversionId: string): Promise<Conversion> => {
  const response = await fetch(`${API_BASE_URL}/${ENDPOINT_URL}/${conversionId}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  if (!response.ok) {
    throw new Error('Failed to fetch schema')
  }
  return response.json()
}
