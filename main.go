package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ez-connect/webhook/internal"
	"github.com/ez-connect/webhook/pkg/core"
)

// Args passed via build in the `Makefile`
// -ldflags="-X 'pkg/main.BuildDate=$(name)' -X 'pkg/main.Branch=$(version)...'"
var (
	Name        string
	Description string
	Version     string
	BuildDate   string
	Branch      string
	Hash        string
	BuildMode   string
)

func main() {
	filename := flag.String("c", "", "path to the configuration file")
	isHelp := flag.Bool("h", false, "show help")
	flag.Usage = usage
	flag.Parse()

	if *isHelp {
		flag.Usage()
	}

	var (
		c   core.Config
		err error
	)

	if *filename == "" {
		log.Printf("no config file provided, using defaults\n")
		c = core.DefaultConfig
	}

	c, err = core.LoadConfig(*filename)
	if err != nil {
		log.Printf("cannot load config: %s error: %s\n", *filename, err)
		c = core.DefaultConfig
	}

	internal.Serve(c)

}

func usage() {
	fmt.Printf("Webhook v%s\n-%s/%s", Version, Hash, Branch)
	fmt.Println("Usage: webhook serve [-c path/to/config/file.jsonc]")
	os.Exit(0)
}
