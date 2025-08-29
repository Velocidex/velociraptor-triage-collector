package main

import (
	"io"
	"log"
	"os"

	"github.com/Velocidex/velociraptor-triage-collector/converters"
	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	uac_cmd      = app.Command("uac", "Convert UAC rules to standard form.")
	uac_filename = uac_cmd.Arg("filename", "UAC file to convert").Required().String()
)

func doUAC() error {
	logger := log.New(io.Discard, "", 0)
	if *debug {
		logger.SetOutput(os.Stderr)
	}

	output, err := converters.UACConvertFile(*uac_filename)
	if err != nil {
		return err
	}

	println(output)
	return nil
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {
		case uac_cmd.FullCommand():
			err := doUAC()
			kingpin.FatalIfError(err, "Coverting UAC artifact")

		default:
			return false
		}
		return true
	})
}
