package model

import "encoding/json"

type HttpResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Task    *Task  `json:"task"`
}

var SUCCESS, _ = json.Marshal(&HttpResponse{true, "", nil})

func Fail(errMsg string) *HttpResponse {
	return &HttpResponse{false, errMsg, nil}
}

func Success(task *Task) *HttpResponse {
	return &HttpResponse{true, "", task}
}
