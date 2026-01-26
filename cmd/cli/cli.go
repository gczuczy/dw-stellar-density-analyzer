package cli

import (
	"os"
	"fmt"
	"github.com/knadh/koanf/v2"

	ds "github.com/gczuczy/dw-stellar-density-analyzer/pkg/densitysurvey"
)

func Run() {

	k := koanf.New(".")

	err := parseArgs(k)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	entry, err := ds.NewEntrySheet(k.String(`sheetid`), k.String(`sa-creds`))
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
