package sdaservice

import (
	"os"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/config"
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/db"
)

func parseArgs(k *koanf.Koanf) error {
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		f.PrintDefaults()
		os.Exit(0)
	}

	f.StringP("sa-creds", "s", "credentials.json", "The Google Service Account credentials json")
	f.StringP("config", "c", "~/.edsda.yaml", "Path to the configuration file")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}

	return k.Load(posflag.Provider(f, ".", k), nil)
}

func Run() {
	var cfg *config.Config

	k := koanf.New(".")
	k.Print()

	err := parseArgs(k)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	if cfg, err = config.ParseConfig(k); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	if err = db.Init(&cfg.DB); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	defer db.Pool.Close()
}
