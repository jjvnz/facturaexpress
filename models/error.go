package models

import (
	"time"
)

type ErrorJson struct {
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Source    string `json:"source"`
}

func (e *ErrorJson) DefaultErrorInit(title, timestamp, message, errorType, source string) {
	e.Title = title
	e.Timestamp = timestamp
	e.Message = message
	e.Type = errorType
	e.Source = source
}

func ErrorResponseInit(title, message string) *ErrorJson {
	var newError ErrorJson
	newError.DefaultErrorInit(
		title,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		message,
		"genericError",
		"internal",
	)
	return &newError
}

func (e *ErrorJson) Error() string {
	return e.Message
}
