package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	cli "github.com/jawher/mow.cli"
)

func main() {
	var fl, fr string
	app := cli.App("compare-flags", "Compare compiler flags")
	app.Spec = "FL FR"
	app.StringArgPtr(&fl, "FL", "", "First set of flags")
	app.StringArgPtr(&fr, "FR", "", "Second set of flags")

	app.Action = func() {

		ml := map[string]bool{}
		mr := map[string]bool{}
		for _, s := range strings.Split(fl, " ") {
			if s != "" {
				ml[s] = true
			}
		}
		for _, s := range strings.Split(fr, " ") {
			if s != "" {
				mr[s] = true
			}
		}

		delete(ml, " ")
		delete(mr, " ")

		for k := range ml {
			_, ok := mr[k]
			if ok {
				delete(ml, k)
				delete(mr, k)
			}
		}

		ssl := []string{}
		ssr := []string{}
		for k := range ml {
			ssl = append(ssl, k)
		}
		for k := range mr {
			ssr = append(ssr, k)
		}
		ssl = sort.StringSlice(ssl)
		ssr = sort.StringSlice(ssr)

		fmt.Println("First:")
		for _, s := range ssl {
			fmt.Println("  ", s)
		}
		fmt.Println("Second:")
		for _, s := range ssr {
			fmt.Println("  ", s)
		}
	}
	app.Run(os.Args)

}
