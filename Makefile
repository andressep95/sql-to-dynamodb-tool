.PHONY: lambda backend localstack localstack-destroy prod prod-destroy docker-up docker-down

docker-up:
	@echo "🐳 Starting LocalStack Pro..."
	@if [ -z "$$LOCALSTACK_AUTH_TOKEN" ] && [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs) && \
		docker-compose up -d; \
	elif [ -n "$$LOCALSTACK_AUTH_TOKEN" ]; then \
		docker-compose up -d; \
	else \
		echo "❌ LOCALSTACK_AUTH_TOKEN not found. Set it in .env or export it."; \
		exit 1; \
	fi
	@echo "⏳ Waiting for LocalStack to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		if docker logs localstack-main 2>&1 | grep -q "Ready."; then \
			break; \
		fi; \
		sleep 3; \
	done
	@docker logs localstack-main 2>&1 | grep -q "Successfully activated" && \
		echo "✅ LocalStack Pro activated" || \
		echo "⚠️  License not activated - check your AUTH_TOKEN"
	@docker logs localstack-main 2>&1 | grep -q "Ready." && \
		echo "✅ LocalStack is ready" || \
		(echo "❌ LocalStack failed to start"; exit 1)

docker-down:
	@echo "🛑 Stopping LocalStack..."
	docker-compose down
	@echo "✅ LocalStack stopped"

lambda:
	@echo "🔨 Building all Lambdas..."
	@for dir in lambda/*; do \
		if [ -f $$dir/go.mod ]; then \
			echo "➡️  Building $$(basename $$dir)"; \
			cd $$dir && \
			GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
			go build -tags lambda.norpc -o bootstrap . && \
			zip -q function.zip bootstrap && \
			rm bootstrap && \
			cd - > /dev/null; \
			echo "✅ $$(basename $$dir) built"; \
		fi \
	done

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
