#!/bin/bash
# ============================================
# Validate Bedrock Model Access
# ============================================

set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
PROFILE="${AWS_PROFILE:-default}"
MODEL_ID="us.anthropic.claude-sonnet-4-20250514-v1:0"

echo "Validating Bedrock access..."
echo "  Region:  $REGION"
echo "  Profile: $PROFILE"
echo "  Model:   $MODEL_ID"
echo ""

# Check AWS credentials
echo "1. Checking AWS credentials..."
aws sts get-caller-identity --profile "$PROFILE" --region "$REGION" || {
  echo "ERROR: AWS credentials not configured. Run 'aws configure --profile $PROFILE'"
  exit 1
}
echo ""

# Check Bedrock model access
echo "2. Checking Bedrock model access..."
aws bedrock get-foundation-model \
  --model-identifier "anthropic.claude-sonnet-4-20250514-v1:0" \
  --profile "$PROFILE" \
  --region "$REGION" \
  --query 'modelDetails.modelId' \
  --output text || {
  echo "ERROR: Cannot access Bedrock model. Ensure the model is enabled in your AWS account."
  exit 1
}
echo ""

# Test invocation with a simple prompt
echo "3. Testing model invocation..."
TMPFILE=$(mktemp)
aws bedrock-runtime invoke-model \
  --model-id "$MODEL_ID" \
  --profile "$PROFILE" \
  --region "$REGION" \
  --content-type "application/json" \
  --cli-binary-format raw-in-base64-out \
  --body '{"anthropic_version":"bedrock-2023-05-31","max_tokens":50,"messages":[{"role":"user","content":"Say hello in one word."}]}' \
  "$TMPFILE" 2>&1 || {
  echo "ERROR: Model invocation failed. Check IAM permissions for bedrock:InvokeModel."
  exit 1
}

RESPONSE=$(cat "$TMPFILE")
rm -f "$TMPFILE"
echo "Response: $RESPONSE"
echo ""
echo "Bedrock validation successful!"
