package httpResponse

import "encoding/json"

type HttpResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var SUCCESS, _ = json.Marshal(&HttpResponse{true, ""})

func Fail(errMsg string) *HttpResponse {
	return &HttpResponse{false, errMsg}
}
