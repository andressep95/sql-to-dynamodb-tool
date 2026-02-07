#!/bin/bash

# Script para crear un SUPER_ADMIN en Cognito
# Uso: ./create-super-admin.sh <email> <password>

set -e

if [ $# -ne 2 ]; then
  echo "Usage: $0 <email> <password>"
  echo "Example: $0 admin@example.com SuperSecure123!"
  exit 1
fi

EMAIL=$1
PASSWORD=$2

# Obtener User Pool ID desde Terraform
cd infra/terraform/environments/prod
USER_POOL_ID=$(terraform output -raw cognito_user_pool_id)
cd -

echo "🔐 Creating SUPER_ADMIN user..."
echo "   Email: $EMAIL"
echo "   User Pool: $USER_POOL_ID"

# Crear usuario
aws cognito-idp admin-create-user \
  --user-pool-id "$USER_POOL_ID" \
  --username "$EMAIL" \
  --user-attributes \
    Name=email,Value="$EMAIL" \
    Name=email_verified,Value=true \
    Name=custom:role,Value=SUPER_ADMIN \
  --message-action SUPPRESS

echo "✅ User created"

# Establecer password permanente
aws cognito-idp admin-set-user-password \
  --user-pool-id "$USER_POOL_ID" \
  --username "$EMAIL" \
  --password "$PASSWORD" \
  --permanent

echo "✅ Password set"
echo ""
echo "🎉 SUPER_ADMIN created successfully!"
echo "   Login with: $EMAIL / $PASSWORD"
