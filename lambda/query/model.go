package main

// V2Response is a clean API Gateway V2 response without MultiValueHeaders/Cookies null fields.
type V2Response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded,omitempty"`
}
