
locals {
  component_name = "lambda-core"

  lambda_configs = {
    healthcheck = {
      description                    = "Healthcheck"
      use_case                       = "UC-HEALTHCHECK-001"
      api_operation                  = "healthcheck"
      memory_size                    = 128
      timeout                        = 5
      reserved_concurrent_executions = 100
      log_retention_days             = 30
    }
  }
}
