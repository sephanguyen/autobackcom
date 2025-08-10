package utils

import "autobackcom/internal/api/dto"

func Success(data interface{}) dto.APIResponse {
	return dto.APIResponse{Status: "ok", Data: data}
}

func Error(msg string) dto.APIResponse {
	return dto.APIResponse{Status: "error", Error: msg}
}
