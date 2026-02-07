package auth

import (
	"github.com/gorilla/mux"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
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
