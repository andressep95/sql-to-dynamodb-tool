# Configuración de Cognito en el Frontend

## Variables de Entorno Requeridas

El frontend necesita dos variables de entorno para conectarse a AWS Cognito:

```bash
VITE_COGNITO_USER_POOL_ID=us-east-1_XXXXXXXXX
VITE_COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Obtener los Valores desde Terraform

### Opción 1: Terraform Output (Recomendado)

```bash
cd infra/terraform/environments/prod
terraform output cognito_user_pool_id
terraform output cognito_client_id
```

### Opción 2: AWS Console

1. Ve a **AWS Cognito** → **User Pools**
2. Selecciona el pool `sql-to-nosql-parser-prod`
3. Copia el **User Pool ID** (formato: `us-east-1_XXXXXXXXX`)
4. Ve a **App Integration** → **App clients**
5. Copia el **Client ID**

## Configuración Local

1. Crea un archivo `.env.local` en la raíz del proyecto frontend:

```bash
cp .env.local.example .env.local
```

2. Edita `.env.local` y reemplaza los valores:

```bash
VITE_BASE_PATH_URL=https://app-sql.cloudcentinel.com
VITE_ENDPOINT_URL=prod/api/v1/schemas
VITE_COGNITO_USER_POOL_ID=us-east-1_ABC123DEF
VITE_COGNITO_CLIENT_ID=1a2b3c4d5e6f7g8h9i0j
```

3. Reinicia el servidor de desarrollo:

```bash
npm run dev
```

## Flujo de Autenticación

### 1. Registro de Usuario

- El usuario ingresa email y contraseña
- Cognito envía un código de verificación al email
- El usuario confirma con el código recibido

### 2. Login

- El usuario ingresa email y contraseña
- Cognito valida las credenciales
- Se obtiene un JWT token (ID Token)
- El token se almacena en localStorage (manejado por `amazon-cognito-identity-js`)

### 3. Llamadas API Autenticadas

Todas las llamadas a la API incluyen automáticamente el header:

```
Authorization: Bearer <JWT_TOKEN>
```

El token se obtiene dinámicamente en cada request mediante `getIdToken()`.

### 4. Validación en API Gateway

API Gateway valida el JWT contra Cognito:

- Verifica la firma del token
- Valida el issuer (User Pool)
- Valida el audience (Client ID)
- Verifica que no haya expirado

### 5. Tenant Isolation

El backend extrae el `tenant_id` del JWT y lo usa para:

- Filtrar datos por tenant en DynamoDB
- Asegurar que cada usuario solo vea sus propias conversiones

## Estructura del JWT

El ID Token de Cognito contiene:

```json
{
  "sub": "uuid-del-usuario",
  "email": "usuario@ejemplo.com",
  "cognito:username": "usuario@ejemplo.com",
  "custom:tenant_id": "tenant-uuid",
  "iss": "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXX",
  "aud": "client-id",
  "exp": 1234567890
}
```

## Troubleshooting

### Error: "User pool client does not exist"

- Verifica que `VITE_COGNITO_CLIENT_ID` sea correcto
- Asegúrate de usar el Client ID, no el User Pool ID

### Error: "Invalid issuer"

- Verifica que `VITE_COGNITO_USER_POOL_ID` sea correcto
- El formato debe ser `us-east-1_XXXXXXXXX`

### El token no se envía en las requests

- Verifica que el usuario esté autenticado: `auth.isAuthenticated === true`
- Revisa la consola del navegador para errores de `getIdToken()`

### Redirect loop entre login y home

- Verifica que `checkAuth()` se ejecute correctamente en `main.ts`
- Revisa que el router guard esté configurado en `router/index.ts`

## Testing

### Crear un usuario de prueba

```bash
aws cognito-idp sign-up \
  --client-id <CLIENT_ID> \
  --username test@example.com \
  --password TestPassword123! \
  --user-attributes Name=email,Value=test@example.com

aws cognito-idp admin-confirm-sign-up \
  --user-pool-id <USER_POOL_ID> \
  --username test@example.com
```

### Verificar el token en jwt.io

1. Abre las DevTools del navegador
2. Ve a Application → Local Storage
3. Busca la key que contiene `idToken`
4. Copia el valor y pégalo en https://jwt.io
5. Verifica que el `iss` y `aud` coincidan con tu configuración
