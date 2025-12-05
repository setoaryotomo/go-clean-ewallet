package models

type RequestID struct {
	ID int `json:"id" validate:"required"`
}
