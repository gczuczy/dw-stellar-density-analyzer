package http

import (
	"os"
	"fmt"
	"context"

	"net/http"
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/config"
)

type HTTPService struct {
	config *config.HTTPConfig
	srv http.Server
}

func New(cfg *config.HTTPConfig) (*HTTPService, error) {

	hs := HTTPService{
		config: cfg,
		srv: http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
		},
	}

	return &hs, nil
}

func (hs *HTTPService) Run() error {

	go func() {
		err := hs.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "http.Serve(): %v\n", err)
			os.Exit(2)
		}
	}()
	return nil
}

func (hs *HTTPService) Shutdown() error {
	ctx := context.Background()

	return hs.srv.Shutdown(ctx)
}

func (hs *HTTPService) Close() error {
	return hs.srv.Close()
}
