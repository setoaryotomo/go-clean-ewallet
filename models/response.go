package models

type Response struct {
	StatusCode       string      `json:"statusCode"`
	Success          bool        `json:"success"`
	ResponseDatetime string      `json:"responseDatetime"`
	Result           interface{} `json:"result"`
	Message          string      `json:"message"`
}
