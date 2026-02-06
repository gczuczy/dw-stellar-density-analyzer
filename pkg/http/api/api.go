package api

import (
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/config"
)

var (
	oauth2config *config.OAuth2Config
	scopes string
)

type Response struct {
	Status string `json:"status"`
	Message string `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func Init(r *mux.Router, cfg *config.OAuth2Config) error {
	oauth2config = cfg
	scopes = "openid profile email auth"

	r.HandleFunc("/oauth/config", configHandler)

	return nil
}

func success(r any) any {
	return Response{
		Status: "success",
		Data: r,
	}
}

func returnJson(resp any, code int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	if encdata, err := json.Marshal(resp); err == nil {
		w.WriteHeader(code)
		w.Write(encdata)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
