package google

import (
	"fmt"
	"context"
	"google.golang.org/api/sheets/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type GSpreadsheetsService struct {
	Token *oauth2.Token
	SheetsService *sheets.Service
}

type GSpreadsheet struct {
	Sheet *sheets.Spreadsheet
	SheetsService *sheets.Service
	ID string
}

func NewSheets(credfile string) (*GSpreadsheetsService, error) {
	var err error
	ctx := context.Background()

	gs := &GSpreadsheetsService{}

	creds := option.WithCredentialsFile(credfile)
	scopes := option.WithScopes(sheets.SpreadsheetsScope)

	if gs.SheetsService, err = sheets.NewService(ctx, creds, scopes); err != nil {
		return nil, err
	}

	return gs, nil
}

func (s *GSpreadsheetsService) Sheet(id string) (*GSpreadsheet, error) {
	var err error
	sheet := &GSpreadsheet{
		ID: id,
		SheetsService: s.SheetsService,
	}

	sheet.Sheet, err = s.SheetsService.Spreadsheets.Get(id).Do()
	if err != nil {
		return nil, err
	}

	return sheet, nil
}

func (s *GSpreadsheet) GetSheets() []*sheets.Sheet {
	return s.Sheet.Sheets
}

func (s *GSpreadsheet) ReadRange(sheet string, start string, end string) (*sheets.ValueRange, error) {
	rangestr := fmt.Sprintf("%s!%s:%s", sheet, start, end)
	return s.SheetsService.Spreadsheets.Values.Get(s.ID, rangestr).Do()
}
