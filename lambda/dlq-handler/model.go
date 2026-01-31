package main

// SQSMessageBody represents the message body sent from process_handler via SQS.
type SQSMessageBody struct {
	ConversionID     string `json:"conversionId"`
	SQLContent       string `json:"sqlContent"`
	OptimizationType string `json:"optimizationType"`
	TablesExtracted  int    `json:"tablesExtracted"`
}
