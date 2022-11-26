package toolchain

import (
	"fmt"
	"io"
)

// Chain contains all the information discovered about a compiler
type Chain struct {
	Compiler            string `json:"compiler" yaml:"compiler"` // MSVC/GCC/CLANG
	FullVersion         string `json:"full-version,omitempty" yaml:"full-version,omitempty"`
	Version             string `json:"version,omitempty" yaml:"version,omitempty"`
	Target              string `json:"target,omitempty" yaml:"target,omitempty"`
	ThreadModel         string `json:"thread-model,omitempty" yaml:"thread-model,omitempty"`
	InstalledDir        string `json:"installed-dir,omitempty" yaml:"installed-dir,omitempty"`
	VisualStudioID      string `json:"msvc-id,omitempty" yaml:"msvc-id,omitempty"`
	VisualStudioArch    string `json:"msvc-arch,omitempty" yaml:"msvc-arch,omitempty"`
	VisualStudioVersion string `json:"msvc-version,omitempty" yaml:"msvc-version,omitempty"`
	WindowsSDKVersion   string `json:"windows-sdk,omitempty" yaml:"windows-sdk,omitempty"`
	UCRTVersion         string `json:"ucrt,omitempty" yaml:"ucrt,omitempty"`
	ToolsetVersion      string `json:"toolset,omitempty" yaml:"toolset,omitempty"`

	Tools Toolset `json:"tools" yaml:"tools"` // paths to tool executables

	CCIncludeDirs  []string `json:"cc-include-dirs,omitempty" yaml:"cc-include-dirs,omitempty"`
	CXXIncludeDirs []string `json:"cxx-include-dirs,omitempty" yaml:"cxx-include-dirs,omitempty"`
	LibraryDirs    []string `json:"library-dirs,omitempty" yaml:"library-dirs,omitempty"`
	Environment    []string `json:"environment" yaml:"environment"`
}

func (tc *Chain) PrintSummary(w io.Writer) {
	tgt := tc.Target
	if tc.VisualStudioArch != "" {
		tgt += "." + tc.VisualStudioArch
	}
	ver := tc.Version
	if tc.VisualStudioVersion != "" {
		ver = fmt.Sprintf("%s (%s)", tc.VisualStudioVersion, tc.Version)
	}
	fmt.Fprintf(w, "%s %s targeting '%s'\n", tc.Compiler, ver, tgt)
	cc := tc.Tools[CCompiler]
	cxx := tc.Tools[CXXCompiler]
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
