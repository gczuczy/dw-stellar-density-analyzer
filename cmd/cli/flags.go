package cli

import (
	"os"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"
)

func parseArgs(k *koanf.Koanf) error {

	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		f.PrintDefaults()
		os.Exit(0)
	}

	f.StringP("sa-creds", "s", "credentials.json", "The Google Service Account credentials json")
	f.StringP("sheetid", "i", "", "The ID of the entrypoint google sheet")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}

	return k.Load(posflag.Provider(f, ".", k), nil)
}
