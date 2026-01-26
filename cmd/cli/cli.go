package cli

import (
	"os"
	"fmt"
	"github.com/knadh/koanf/v2"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/google"
	ds "github.com/gczuczy/dw-stellar-density-analyzer/pkg/densitysurvey"
)

func Run() {

	k := koanf.New(".")

	err := parseArgs(k)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	creds := k.String(`sa-creds`)
	ss, err := google.NewSheets(creds)
	if err != nil {
		fmt.Printf("Credentials error: %s", creds)
		return
	}

	entry, err := ds.NewEntrySheet(k.String(`sheetid`), ss)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ids, err := entry.GetSheetIDs()
	if err != nil && len(ids) == 0 {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("SheetIDs: %v\n", ids)
}
