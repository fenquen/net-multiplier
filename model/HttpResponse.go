package model

import (
	"encoding/json"
	"net-multiplier/client"
)

type HttpResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Task    *client.Task `json:"task"`
}

var SUCCESS, _ = json.Marshal(&HttpResponse{true, "", nil})

func Fail(errMsg string) *HttpResponse {
	return &HttpResponse{false, errMsg, nil}
}

func Success(task *client.Task) *HttpResponse {
	return &HttpResponse{true, "", task}
}
