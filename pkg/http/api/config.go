package api

import (
	"fmt"
	"strings"
	"net/http"
)

type Config struct {
	Issuer string `json:"issuer"`
	ClientID string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	Scope string `json:"scope"`
}

func configHandler(w http.ResponseWriter, r *http.Request) {

	var host string
	if r.Host == "localhost" || r.Host == "127.0.0.1" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1:") {
		host = fmt.Sprintf("http://%s", r.Host)
	} else {
		host = fmt.Sprintf("https://%s", r.Host)
	}

	resp := success(Config{
		Issuer: oauth2config.Issuer,
		ClientID: oauth2config.ClientID,
		RedirectURI: fmt.Sprintf("%s/api/auth/callback", host),
		Scope: scopes,
	})
	returnJson(resp, 200, w)
}
