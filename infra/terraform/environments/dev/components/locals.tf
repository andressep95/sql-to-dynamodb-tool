
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

    conversion_worker = {
      description                    = "conversion_worker"
      use_case                       = "UC-conversion_worker-001"
      api_operation                  = "conversion_worker"
      memory_size                    = 256
      timeout                        = 120
      reserved_concurrent_executions = 5
      log_retention_days             = 30
    }

    query_handler = {
      description                    = "query_handler"
      use_case                       = "UC-query_handler-001"
      api_operation                  = "query_handler"
      memory_size                    = 256
      timeout                        = 120
      reserved_concurrent_executions = 5
      log_retention_days             = 30
    }

    dlq_handler = {
      description                    = "dlq_handler"
      use_case                       = "UC-dlq_handler-001"
      api_operation                  = "dlq_handler"
      memory_size                    = 128
      timeout                        = 30
      reserved_concurrent_executions = 5
      log_retention_days             = 30
    }
  }
}
