package clang

import (
	"fmt"
	"io"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
)

type Installation struct {
	Ver
	CCompiler toolchain.Executable `json:"c-compiler" yaml:"c-compiler"`
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "%s %s\n", i.Ver.Implementation, i.Version)
	fmt.Fprintf(w, "- full version: '%s'\n", i.FullVersion)
	fmt.Fprintf(w, "- primary target: %s\n", i.Target.Original)
	fmt.Fprintf(w, "  - os: %s\n", i.Target.OS)
	fmt.Fprintf(w, "  - arch: %s\n", i.Target.Arch)
	fmt.Fprintf(w, "  - abi: %s\n", i.Target.ABI)
	fmt.Fprintf(w, "  - libc: %s\n", i.Target.LibC)
	fmt.Fprintf(w, "- thread model: %s\n", i.ThreadModel)
	fmt.Fprintf(w, "- CC primary path: '%s'\n", i.CCompiler.PrimaryPath)
	if len(i.CCompiler.Subcommands) > 0 {
		fmt.Fprintf(w, "- CC subcommands: %s\n", strings.Join(i.CCompiler.Subcommands, " "))
	}
	for _, v := range i.CCompiler.OtherPaths {
		fmt.Fprintf(w, "- CC alternative path: '%s'\n", v)
	}
	for _, v := range i.CCompiler.SymLinks {
		fmt.Fprintf(w, "- CC symlink path: '%s'\n", v)
	}
	fmt.Fprintf(w, "- installed dir: %s\n", i.InstalledDir)
}
