import {
  CognitoUserPool,
  CognitoUser,
  AuthenticationDetails,
  type CognitoUserSession,
  type CognitoUserAttribute,
} from 'amazon-cognito-identity-js'

const userPool = new CognitoUserPool({
  UserPoolId: import.meta.env.VITE_COGNITO_USER_POOL_ID || '',
  ClientId: import.meta.env.VITE_COGNITO_CLIENT_ID || '',
})

export function signUp(
  email: string,
  password: string,
): Promise<{ userSub: string }> {
  return new Promise((resolve, reject) => {
    userPool.signUp(email, password, [], [], (err, result) => {
      if (err) return reject(err)
      resolve({ userSub: result!.userSub })
    })
  })
}

export function confirmSignUp(email: string, code: string): Promise<void> {
  const cognitoUser = new CognitoUser({ Username: email, Pool: userPool })
  return new Promise((resolve, reject) => {
    cognitoUser.confirmRegistration(code, true, (err) => {
      if (err) return reject(err)
      resolve()
    })
  })
}

export function signIn(
  email: string,
  password: string,
): Promise<CognitoUserSession> {
  const cognitoUser = new CognitoUser({ Username: email, Pool: userPool })
  const authDetails = new AuthenticationDetails({
    Username: email,
    Password: password,
  })

  return new Promise((resolve, reject) => {
    cognitoUser.authenticateUser(authDetails, {
      onSuccess: (session) => resolve(session),
      onFailure: (err) => reject(err),
    })
  })
}

export function signOut(): void {
  const cognitoUser = userPool.getCurrentUser()
  if (cognitoUser) {
    cognitoUser.signOut()
  }
}

export function getSession(): Promise<CognitoUserSession | null> {
  return new Promise((resolve) => {
    const cognitoUser = userPool.getCurrentUser()
    if (!cognitoUser) return resolve(null)

    cognitoUser.getSession(
      (err: Error | null, session: CognitoUserSession | null) => {
        if (err || !session || !session.isValid()) return resolve(null)
        resolve(session)
      },
    )
  })
}

export async function getIdToken(): Promise<string | null> {
  const session = await getSession()
  return session?.getIdToken().getJwtToken() ?? null
}

export function getCurrentUser(): CognitoUser | null {
  return userPool.getCurrentUser()
}

export function getUserAttributes(): Promise<CognitoUserAttribute[]> {
  return new Promise((resolve, reject) => {
    const cognitoUser = userPool.getCurrentUser()
    if (!cognitoUser) return reject(new Error('No current user'))

    cognitoUser.getSession((err: Error | null) => {
      if (err) return reject(err)

      cognitoUser.getUserAttributes((attrErr, attributes) => {
        if (attrErr) return reject(attrErr)
        resolve(attributes || [])
      })
    })
  })
}
