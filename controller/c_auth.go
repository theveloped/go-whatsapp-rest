package controller

import (
	"encoding/json"
	"net/http"

	svc "github.com/theveloped/go-whatsapp-rest/service"
)

// GetAuth Function to Get Authorization Token
func GetAuth(w http.ResponseWriter, r *http.Request) {
	var reqBody svc.ReqGetBasic
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	if len(reqBody.Username) == 0 || len(reqBody.Password) == 0 {
		svc.ResponseBadRequest(w, "invalid authorization")
		return
	}

	if reqBody.Password != svc.Config.GetString("AUTH_PASSWORD") {
		svc.ResponseBadRequest(w, "invalid authorization")
		return
	}

	token, err := svc.GetJWTToken(reqBody.Username)
	if err != nil {
		svc.ResponseInternalError(w, err.Error())
		return
	}

	var response svc.ResGetJWT

	response.Status = true
	response.Code = http.StatusOK
	response.Message = "Success"
	response.Data.Token = token

	svc.ResponseWrite(w, response.Code, response)
}
