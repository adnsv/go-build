package proj

import (
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
)

// ExtensionType defines to which category a file
// belongs based on file extension.
type ExtensionType int

// Supported ExtensionType values
const (
	ExtensionUnknown = ExtensionType(iota)
	ExtensionHeader
	ExtensionCompilationUnitCXX
	ExtensionCompilationUnitC
	ExtensionCompilationUnitASM
	ExtensionResource
)

var ExtensionTools = map[ExtensionType]toolchain.Tool{
	ExtensionCompilationUnitCXX: toolchain.CXXCompiler,
	ExtensionCompilationUnitC:   toolchain.CCompiler,
	ExtensionCompilationUnitASM: toolchain.ASMCompiler,
	ExtensionResource:           toolchain.ResourceCompiler,
}

func GetExtensionType(ext string) ExtensionType {
	return KnownExtensions[strings.ToLower(ext)]
}

// KnownExtensions maps known file extensions to ExtensionType.
var KnownExtensions = map[string]ExtensionType{
	".h":   ExtensionHeader,
	".hpp": ExtensionHeader,
	".cpp": ExtensionCompilationUnitCXX,
	".cxx": ExtensionCompilationUnitCXX,
	".cc":  ExtensionCompilationUnitCXX,
	".c":   ExtensionCompilationUnitC,
	".asm": ExtensionCompilationUnitASM,
	".s":   ExtensionCompilationUnitASM,
	".rc":  ExtensionResource,
}
