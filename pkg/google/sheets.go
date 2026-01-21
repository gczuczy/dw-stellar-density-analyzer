package google

import (
	"fmt"
	"context"
	"google.golang.org/api/sheets/v4"
	"golang.org/x/oauth2"
	/*
	"github.com/int128/oauth2cli"
	g "golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"github.com/pkg/browser"
	*/
)

type GSheets struct {
	Token *oauth2.Token
	SheetsService *sheets.Service
}

func NewSheets() (*GSheets, error) {
	var err error
	ctx := context.Background()
	fmt.Printf("lofasz\n")

	gs := &GSheets{}

	/*
	pkceVerifier := oauth2.GenerateVerifier()
	ready := make(chan string, 1)
	cliCfg := oauth2cli.Config{
		OAuth2Config: oauth2.Config{
			Endpoint: g.Endpoint,
		},
		LocalServerSuccessHTML: `Great Success Great Leader!`,
		AuthCodeOptions:      []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(pkceVerifier)},
		TokenRequestOptions:  []oauth2.AuthCodeOption{oauth2.VerifierOption(pkceVerifier)},
		LocalServerReadyChan: ready,
		Logf: func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
			fmt.Printf("\n")
		},
	}
	cliCfg.OAuth2Config.ClientID = "32555940559.apps.googleusercontent.com"
	cliCfg.OAuth2Config.Scopes = []string{"email"}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-ready:
			fmt.Printf("Open %s\n", url)
			if err := browser.OpenURL(url); err != nil {
				return err
			}
			return nil
		case <- ctx.Done():
			return fmt.Errorf("Context done while waiting for auth: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		if token, err := oauth2cli.GetToken(ctx, cliCfg); err != nil {
			return err
		} else {
			gs.Token = token
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	*/

	if gs.SheetsService, err = sheets.NewService(ctx); err != nil {
		return nil, err
	}

	return gs, nil
}
