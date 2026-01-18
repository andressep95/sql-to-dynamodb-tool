.PHONY: lambda backend localstack localstack-destroy prod prod-destroy 

lambda:
	@echo "🔨 Building diagrams Lambda..."
	cd lambda/diagrams && \
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap . && \
	zip -q function.zip bootstrap && \
	rm bootstrap
	@echo "✅ diagrams built"

## SOLO PARA DESARROLLO
localstack:
	@echo "🔨 Starting localstack..."
	cd infra/terraform/environments/dev && \
	terraform init && terraform apply -auto-approve
	@echo "✅ Localstack started"

localstack-destroy:
	@echo "🧹 Destroying LocalStack environment..."
	cd infra/terraform/environments/dev && \
	terraform destroy -auto-approve
	@echo "✅ LocalStack destroyed"

## SOLO PARA PRODUCCION

backend:
	@echo "🔨 Initializing backend..."
	cd infra/terraform/backend && \
	terraform init && terraform apply -auto-approve
	@echo "✅ Backend initialized"

prod:
	@echo "🔨 Deploying to production..."
	cd infra/terraform/environments/prod && \
	terraform init && terraform apply -auto-approve
	@echo "✅ Production deployed"

prod-destroy:
	@echo "⚠️  Destroying production environment..."
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	cd infra/terraform/environments/prod && \
	terraform destroy -auto-approve
	@echo "✅ Production destroyed"
