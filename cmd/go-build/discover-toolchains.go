package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/adnsv/go-build/compiler"
	"github.com/alecthomas/kong"
)

type DiscoverToolchains struct {
	Verbose       bool     `short:"v" help:"Show verbose output"`
	Output        string   `short:"o" type:"path" help:"Write output to the specified file"`
	Format        string   `short:"f" enum:"summary,json" placeholder:"summary|json" default:"summary" help:"Output format (defaults to summary)"`
	Types         []string `short:"t" enum:"msvc,clang,gcc" help:"Comma separated toolchain types (msvc|clang|gcc)"`
	Installations bool     `short:"i" help:"Show compiler installations instead of toolchains"`
}

func (cmd *DiscoverToolchains) Run(ctx *kong.Context) error {
	var buf []byte
	var err error
	var feedback func(string)
	if cmd.Verbose {
		feedback = func(s string) {
			log.Println(s)
		}
	}
	if cmd.Installations {
		ii := compiler.DiscoverInstallations(cmd.Types, feedback)
		if cmd.Format == "json" {
			buf, err = json.MarshalIndent(ii, "", "  ")
			if err != nil {
				return err
			}
		} else if cmd.Format == "summary" {
			w := &bytes.Buffer{}
			for _, i := range ii {
				i.PrintSummary(w)
			}
			if len(ii) == 0 {
				fmt.Fprintln(w, "no installations found")
			}
			buf = w.Bytes()
		} else {
			return fmt.Errorf("unsupported format '%s'", cmd.Format)
		}
	} else {
		tt := compiler.DiscoverToolchains(true, cmd.Types, feedback)
		if cmd.Format == "json" {
			buf, err = json.MarshalIndent(tt, "", "  ")
			if err != nil {
				return err
			}
		} else if cmd.Format == "summary" {
			w := &bytes.Buffer{}
			for _, tc := range tt {
				tc.PrintSummary(w)
			}
			if len(tt) == 0 {
				fmt.Fprintln(w, "no compilers found")
			}
			buf = w.Bytes()
		} else {
			return fmt.Errorf("unsupported format '%s'", cmd.Format)
		}

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
		return err
	}
	return err
}
