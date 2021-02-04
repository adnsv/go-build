package gcc

import (
	"fmt"
	"io"

	"github.com/adnsv/go-build/compiler/toolchain"
)

// Ver is a version info extracted from `gcc -v` output
type Ver struct {
	FullVersion string   `json:"full-version"`
	Version     string   `json:"version"`
	Target      string   `json:"target"`
	ThreadModel string   `json:"thread-model"`
	IncludeDirs []string `json:"include-dirs"`
}

type Installation struct {
	Ver
	CCompiler toolchain.Executable `json:"c-compiler"`
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "GCC %s\n", i.Version)
	fmt.Fprintf(w, "- full version: '%s'\n", i.FullVersion)
	fmt.Fprintf(w, "- target: %s\n", i.Target)
	fmt.Fprintf(w, "- thread model: %s\n", i.ThreadModel)
	fmt.Fprintf(w, "- CC primary path: '%s'\n", i.CCompiler.PrimaryPath)
	for _, v := range i.CCompiler.OtherPaths {
		fmt.Fprintf(w, "- CC alternative path: '%s'\n", v)
	}
	for _, v := range i.CCompiler.SymLinks {
		fmt.Fprintf(w, "- CC symlink path: '%s'\n", v)
	}
}
