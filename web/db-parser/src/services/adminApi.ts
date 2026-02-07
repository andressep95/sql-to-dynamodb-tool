import { getIdToken } from './authService'

const API_BASE_URL = import.meta.env.VITE_BASE_PATH_URL || ''
const API_VERSION = 'prod/api/v1'

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
    const message = errorBody?.message || getErrorMessage(response.status)
    throw new Error(message)
  }
  
  return response.json()
}

function getErrorMessage(status: number): string {
  switch (status) {
    case 400: return 'Solicitud inválida'
    case 403: return 'No tienes permisos para realizar esta acción'
    case 404: return 'Recurso no encontrado'
    case 500: return 'Error interno del servidor'
    case 503: return 'Servicio no disponible'
    default: return 'Error en la conexión con el servidor'
  }
}

export interface User {
  sub: string
  email: string
  role: string
  tenantId: string
  enabled: boolean
}

export interface CreateUserRequest {
  email: string
  password: string
  role: string
  tenantId: string
}

export interface Tenant {
  id: string
  name: string
  description: string
  userCount: number
}

export interface CreateTenantRequest {
  name: string
  description: string
}

export interface CreateInvitationRequest {
  tenantId: string
  role: string
  email?: string
}

export interface InvitationResponse {
  code: string
  tenantId: string
  role: string
  expiresAt: number
  createdBy: string
}

export interface RegisterRequest {
  invitationCode: string
  email: string
  password: string
}

export const getUsers = async (): Promise<{ users: User[]; count: number }> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/users`, {
      method: 'GET',
      headers: await getHeaders(),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const createUser = async (user: CreateUserRequest): Promise<any> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/users`, {
      method: 'POST',
      headers: await getHeaders(),
      body: JSON.stringify(user),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const getTenants = async (): Promise<{ tenants: Tenant[]; count: number }> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/tenants`, {
      method: 'GET',
      headers: await getHeaders(),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const createTenant = async (tenant: CreateTenantRequest): Promise<any> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/tenants`, {
      method: 'POST',
      headers: await getHeaders(),
      body: JSON.stringify(tenant),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const createInvitation = async (invitation: CreateInvitationRequest): Promise<InvitationResponse> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/invitations`, {
      method: 'POST',
      headers: await getHeaders(),
      body: JSON.stringify(invitation),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const getInvitation = async (code: string): Promise<InvitationResponse> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/invitations/${code}`, {
      method: 'GET',
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}

export const register = async (data: RegisterRequest): Promise<any> => {
  try {
    const response = await fetch(`${API_BASE_URL}/${API_VERSION}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    return handleResponse(response)
  } catch (error: any) {
    if (error.message === 'Failed to fetch') {
      throw new Error('No se pudo conectar con el servidor. Verifica tu conexión a internet.')
    }
    throw error
  }
}
