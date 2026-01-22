package cli

import (
	"os"
	"fmt"
	"github.com/knadh/koanf/v2"
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/google"
)

func Run() {

	k := koanf.New(".")

	err := parseArgs(k)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	k.Print()

	ss, err := google.NewSheets(k.String(`sa-creds`))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	s, err := ss.Sheet(k.String(`sheetid`))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	for n, sheet := range s.GetSheets() {
		fmt.Printf("Sheet %d: %s\n", n, sheet.Properties.Title)
	}
}
