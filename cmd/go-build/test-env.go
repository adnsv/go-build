package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/adnsv/go-build/compiler/cc"
	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

type TestEnv struct {
	Output string `short:"o" type:"path" help:"Write output to the specified file"`
	Format string `short:"f" enum:"summary,json,yaml" placeholder:"summary|json|yaml" default:"summary" help:"Output format (defaults to summary)"`
}

func (cmd *TestEnv) Run(ctx *kong.Context) error {
	b, err := cc.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	var buf []byte
	switch cmd.Format {
	case "json":
		buf, err = json.MarshalIndent(b.Core, "", "  ")
	case "yaml":
		buf, err = yaml.Marshal(b.Core)
	case "summary":
		w := &bytes.Buffer{}
		b.Core.PrintSummary(w)
		buf = w.Bytes()
	default:
		return fmt.Errorf("unsupported format '%s'", cmd.Format)
	}
	if err != nil {
		return err
	}

	if cmd.Output == "" {
		_, err = os.Stdout.Write(buf)
	} else {
		fmt.Fprintf(os.Stderr, "writing results to %s ... ", cmd.Output)
		err = os.WriteFile(cmd.Output, buf, 0666)
		if err == nil {
			fmt.Fprintf(os.Stderr, "SUCCEEDED\n")
		} else {
			fmt.Fprintf(os.Stderr, "FAILED\n")
		}
	}
	return err
}
