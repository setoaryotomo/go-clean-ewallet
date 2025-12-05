package models

import "time"

type Response struct {
	StatusCode 		 string `json:"statusCode"`
	Success			 bool	`json:"success"`
	ResponseDatetime time.Time `json:"responseDatetime"`
	Result			 interface{} `json:"result"`
	Message			 string	`json:"message"`
}