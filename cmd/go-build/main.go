package main

import "github.com/alecthomas/kong"

var cli struct {
	DiscoverToolchains DiscoverToolchains `cmd:"" help:"Show available C/C++ toolchains."`
	Version            kong.VersionFlag   `short:"v" help:"Print version information and quit."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("go-build"),
		kong.Description("GO build utilitity."),
		kong.UsageOnError(),
		kong.Vars{"version": app_version()},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
