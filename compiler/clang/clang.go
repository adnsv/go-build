package clang

import (
	"fmt"
	"io"

	"github.com/adnsv/go-build/compiler/toolchain"
)

// Ver is a version info extracted from `clang -v` output
type Ver struct {
	FullVersion    string   `json:"full-version" yaml:"full-version"`
	Version        string   `json:"version" yaml:"version"`
	Target         string   `json:"target" yaml:"target"`
	ThreadModel    string   `json:"thread-model" yaml:"thread-model"`
	CCIncludeDirs  []string `json:"cc-include-dirs" yaml:"cc-include-dirs"`
	CXXIncludeDirs []string `json:"cxx-include-dirs" yaml:"cxx-include-dirs"`
	InstalledDir   string   `json:"installed-dir" yaml:"installed-dir"`
}

type Installation struct {
	Ver
	CCompiler toolchain.Executable `json:"c-compiler" yaml:"c-compiler"`
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "CLANG %s\n", i.Version)
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
	fmt.Fprintf(w, "- installed dir: %s\n", i.InstalledDir)
}
