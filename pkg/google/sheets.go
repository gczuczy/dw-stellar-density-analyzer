package google

import (
	_ "fmt"
	"context"
	"google.golang.org/api/sheets/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	/*
	"github.com/int128/oauth2cli"
	g "golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"github.com/pkg/browser"
	*/
)

type GSheetsService struct {
	Token *oauth2.Token
	SheetsService *sheets.Service
}

type GSheet struct {
	Sheet *sheets.Spreadsheet
	ID string
}

func NewSheets(credfile string) (*GSheetsService, error) {
	var err error
	ctx := context.Background()

	gs := &GSheetsService{}

	creds := option.WithCredentialsFile(credfile)
	scopes := option.WithScopes(sheets.SpreadsheetsScope)

	if gs.SheetsService, err = sheets.NewService(ctx, creds, scopes); err != nil {
		return nil, err
	}

	return gs, nil
}

func (s *GSheetsService) Sheet(id string) (*GSheet, error) {
	var err error
	sheet := &GSheet{
		ID: id,
	}

	sheet.Sheet, err = s.SheetsService.Spreadsheets.Get(id).Do()
	if err != nil {
		return nil, err
	}
	
	return sheet, nil
}

func (s *GSheet) GetSheets() []*sheets.Sheet {
	return s.Sheet.Sheets
}
