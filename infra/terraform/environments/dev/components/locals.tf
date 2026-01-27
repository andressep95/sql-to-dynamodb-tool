
locals {
  component_name = "lambda-core"

  lambda_configs = {
    process_handler = {
      description                    = "process_handler"
      use_case                       = "UC-process_handler-001"
      api_operation                  = "process_handler"
      memory_size                    = 128
      timeout                        = 5
      reserved_concurrent_executions = 100
      log_retention_days             = 30
    }
  }
}
