import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import {
  signIn as cognitoSignIn,
  signUp as cognitoSignUp,
  confirmSignUp as cognitoConfirmSignUp,
  signOut as cognitoSignOut,
  getSession,
  getUserAttributes,
} from '@/services/authService'

export const useAuthStore = defineStore('auth', () => {
  const isAuthenticated = ref(false)
  const userEmail = ref('')
  const tenantId = ref('')
  const loading = ref(false)

  const displayName = computed(() => {
    if (!userEmail.value) return ''
    return userEmail.value.split('@')[0]
  })

  async function checkAuth() {
    loading.value = true
    try {
      const session = await getSession()
      if (session && session.isValid()) {
        isAuthenticated.value = true
        const attrs = await getUserAttributes()
        for (const attr of attrs) {
          if (attr.getName() === 'email') userEmail.value = attr.getValue()
          if (attr.getName() === 'custom:tenant_id') tenantId.value = attr.getValue()
        }
      } else {
        isAuthenticated.value = false
        userEmail.value = ''
        tenantId.value = ''
      }
    } catch {
      isAuthenticated.value = false
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, password: string) {
    loading.value = true
    try {
      await cognitoSignIn(email, password)
      isAuthenticated.value = true
      userEmail.value = email
      const attrs = await getUserAttributes()
      for (const attr of attrs) {
        if (attr.getName() === 'custom:tenant_id') tenantId.value = attr.getValue()
      }
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string) {
    loading.value = true
    try {
      await cognitoSignUp(email, password)
    } finally {
      loading.value = false
    }
  }

  async function confirmRegistration(email: string, code: string) {
    loading.value = true
    try {
      await cognitoConfirmSignUp(email, code)
    } finally {
      loading.value = false
    }
  }

  function logout() {
    cognitoSignOut()
    isAuthenticated.value = false
    userEmail.value = ''
    tenantId.value = ''
  }

  return {
    isAuthenticated,
    userEmail,
    tenantId,
    loading,
    displayName,
    checkAuth,
    login,
    register,
    confirmRegistration,
    logout,
  }
})
