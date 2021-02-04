package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/adnsv/go-build/compiler"
	cli "github.com/jawher/mow.cli"
)

var check = func(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := cli.App("go-build", "GO build library utilities")

	app.Command("discover-toolchains", "Show available C/C++ toolchains", func(c *cli.Cmd) {
		c.Spec = "[--verbose] [--output=<filename>] [--format=<content-format>] [--types=<type-filter>] [--installations]"
		verbose := c.BoolOpt("v verbose", false, "show verbose output")
		output := c.StringOpt("o output", "", "output file (uses stdout if not specified)")
		format := c.StringOpt("f format", "summary", "output format")
		types := c.StringOpt("t types", "", "comma separated toolchain types (msvc,clang,gcc)")
		installations := c.BoolOpt("i installations", false, "show compiler installations instead of toolchains")

		c.Action = func() {
			buf := []byte{}
			var err error
			var feedback func(string)
			if *verbose {
				feedback = func(s string) {
					log.Println(s)
				}
			}
			if *installations {
				ii := compiler.DiscoverInstallations(strings.Split(*types, ","), feedback)
				if *format == "json" {
					buf, err = json.MarshalIndent(ii, "", "  ")
					check(err)
				} else {
					w := &bytes.Buffer{}
					for _, i := range ii {
						i.PrintSummary(w)
					}
					if len(ii) == 0 {
						fmt.Fprintln(w, "no installations found")
					}
					buf = w.Bytes()
				}
			} else {
				tt := compiler.DiscoverToolchains(true, strings.Split(*types, ","), feedback)
				if *format == "json" {
					buf, err = json.MarshalIndent(tt, "", "  ")
				} else if *format == "summary" {
					w := &bytes.Buffer{}
					for _, tc := range tt {
						tc.PrintSummary(w)
					}
					if len(tt) == 0 {
						fmt.Fprintln(w, "no compilers found")
					}
					buf = w.Bytes()
				} else {
					err = fmt.Errorf("unsupported format '%s'", *format)
				}
				check(err)
			}
			if *output == "" {
				fmt.Println(string(buf))
			} else {
				check(ioutil.WriteFile(*output, buf, 0666))
			}
		}
	})

	app.Run(os.Args)
}
