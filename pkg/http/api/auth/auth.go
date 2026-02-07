package auth

import (
	"github.com/gorilla/mux"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/config"
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/http/wrappers"
)

var (
	oauth2config *config.OAuth2Config
	scopes string
)


func Init(r *mux.Router, cfg *config.OAuth2Config) error {
	oauth2config = cfg
	scopes = "auth"

 	r.HandleFunc("/config", wrappers.Wrap(configHandler))

	return nil
}
