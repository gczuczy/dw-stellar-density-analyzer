package auth

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/http/wrappers"
)

type Config struct {
	Issuer string `json:"issuer"`
	ClientID string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	AuthURL string `json:"authUrl"`
	TokenURL string `json:"tokenUrl"`
	Scope string `json:"scope"`
}

func configHandler(r *http.Request) (any, wrappers.Error) {

	var host string
	if r.Host == "localhost" || r.Host == "127.0.0.1" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1:") {
		host = fmt.Sprintf("http://%s", r.Host)
	} else {
		host = fmt.Sprintf("https://%s", r.Host)
	}

	return Config{
		Issuer: oauth2config.Issuer,
		ClientID: oauth2config.ClientID,
		RedirectURI: fmt.Sprintf("%s/api/auth/callback", host),
		Scope: scopes,
		AuthURL: oauth2config.AuthorizeURL,
		TokenURL: oauth2config.TokenURL,
	}, nil
}
