package toolchain

import (
	"fmt"
	"io"

	"github.com/adnsv/go-build/compiler/triplet"
)

// Chain contains all the information discovered about a compiler
type Chain struct {
	Compiler            string       `json:"compiler" yaml:"compiler"`                                 // msvc|gcc|clang
	Implementation      string       `json:"implementation,omitempty" yaml:"implementation,omitempty"` // msvc|gcc|clang|apple-clang|emscripten|...
	FullVersion         string       `json:"full-version,omitempty" yaml:"full-version,omitempty"`
	Version             string       `json:"version,omitempty" yaml:"version,omitempty"`
	Target              triplet.Full `json:"target,omitempty" yaml:"target,omitempty"`
	ThreadModel         string       `json:"thread-model,omitempty" yaml:"thread-model,omitempty"`
	InstalledDir        string       `json:"installed-dir,omitempty" yaml:"installed-dir,omitempty"`
	VisualStudioID      string       `json:"msvc-id,omitempty" yaml:"msvc-id,omitempty"`
	VisualStudioArch    string       `json:"msvc-arch,omitempty" yaml:"msvc-arch,omitempty"`
	VisualStudioVersion string       `json:"msvc-version,omitempty" yaml:"msvc-version,omitempty"`
	WindowsSDKVersion   string       `json:"windows-sdk,omitempty" yaml:"windows-sdk,omitempty"`
	UCRTVersion         string       `json:"ucrt,omitempty" yaml:"ucrt,omitempty"`
	ToolsetVersion      string       `json:"toolset,omitempty" yaml:"toolset,omitempty"`

	Tools Toolset `json:"tools" yaml:"tools"` // paths to tool executables

	CCIncludeDirs  []string `json:"cc-include-dirs,omitempty" yaml:"cc-include-dirs,omitempty"`
	CXXIncludeDirs []string `json:"cxx-include-dirs,omitempty" yaml:"cxx-include-dirs,omitempty"`
	LibraryDirs    []string `json:"library-dirs,omitempty" yaml:"library-dirs,omitempty"`
	Environment    []string `json:"environment" yaml:"environment"`
}

func (tc *Chain) PrintSummary(w io.Writer) {
	ver := tc.Version
	if tc.VisualStudioVersion != "" {
		ver = fmt.Sprintf("%s (%s)", tc.VisualStudioVersion, tc.Version)
	}
	fmt.Fprintf(w, "%s %s\n", tc.Compiler, ver)
	fmt.Fprintf(w, "- target: %s\n", tc.Target.Original)
	fmt.Fprintf(w, "  - os: %s\n", tc.Target.OS)
	fmt.Fprintf(w, "  - arch: %s\n", tc.Target.Arch)
	fmt.Fprintf(w, "  - abi: %s\n", tc.Target.ABI)
	fmt.Fprintf(w, "  - libc: %s\n", tc.Target.LibC)
	cc, cxx := tc.GetCompilerPaths()
	if cc == cxx {
		fmt.Fprintf(w, "  - C/C++ path: '%s'\n", cc)
	} else {
		if cc != "" {
			fmt.Fprintf(w, "  - C path: '%s'\n", cc)
		}
		if cxx != "" {
			fmt.Fprintf(w, "  - C++ path: '%s'\n", cxx)
		}
	}
}

func (tc *Chain) GetCompilerPaths() (cc, cxx string) {
	if v := tc.Tools[CCompiler]; v != "" {
		cc = v.Path()
	}
	if v := tc.Tools[CXXCompiler]; v != "" {
		cxx = v.Path()
	}
	return
}
