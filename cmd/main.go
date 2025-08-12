package main

import (
	"io"
	"log"
	"os"

	"github.com/Velocidex/velociraptor-triage-collector/compiler"
	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	app = kingpin.New("velotriage",
		"A tool for creating Velotriage Triage artifacts.")

	compile_cmd = app.Command("compile", "Compile all the rules into one rule.")
	config      = compile_cmd.Flag("config", "Config file to use").Required().ExistingFile()
	debug       = app.Flag("debug", "Print more details").Short('v').Bool()

	command_handlers []CommandHandler

	allowed_additional_fields = []string{"details", "vql", "vql_args", "enrichment"}
)

type CommandHandler func(command string) bool

func main() {
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	for _, handler := range command_handlers {
		if handler(command) {
			break
		}
	}
}

func doCompile() error {
	logger := log.New(io.Discard, "", 0)
	if *debug {
		logger.SetOutput(os.Stderr)
	}

	compiler, err := compiler.NewCompiler(*config, logger)
	if err != nil {
		return err
	}

	return compiler.Run()
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {
		case compile_cmd.FullCommand():
			err := doCompile()
			kingpin.FatalIfError(err, "Compiling artifact")

		default:
			return false
		}
		return true
	})
}
