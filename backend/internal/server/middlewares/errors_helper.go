package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"

	"omar-kada/air-compose/api"
)

func sendError(w http.ResponseWriter, errCode api.ErrorCode) {
	sendErrorMessage(w, errCode, "")
}

func sendErrorMessage(w http.ResponseWriter, errCode api.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	builder := strings.Builder{}
	json.NewEncoder(&builder).Encode(api.Error{
		Code:    errCode,
		Message: message,
	})
	http.Error(w, builder.String(), errorCodeToHTTPCode(errCode))
}

func errorCodeToHTTPCode(errCode api.ErrorCode) int {
	switch errCode {
	case api.ErrorCodeINVALIDTOKEN:
		return http.StatusUnauthorized
	case api.ErrorCodeINVALIDCREDENTIALS:
		return http.StatusUnauthorized
	case api.ErrorCodeINVALIDREQUEST:
		return http.StatusBadRequest
	case api.ErrorCodeNOTALLOWED:
		return http.StatusMethodNotAllowed
	case api.ErrorCodeDISABLED:
		return http.StatusMethodNotAllowed
	case api.ErrorCodeNOTFOUND:
		return http.StatusNotFound
	case api.ErrorCodeSERVERERROR:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
