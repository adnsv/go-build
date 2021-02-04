package toolchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Tool specifies a tool inside a toolchain by its function
type Tool int

// Supported tool values
const (
	UnknownTool = Tool(iota)
	CXXCompiler
	CCompiler
	ASMCompiler
	ResourceCompiler
	Archiver
	DLLLinker
	EXELinker

	OBJCopy
	OBJDump
	Runlib
	Strip
)

// Toolset is a collection of tools with filepaths to their executables
type Toolset map[Tool]string

var shortToolNames = map[Tool]string{
	UnknownTool:      "UNKNOWN_TOOL",
	CXXCompiler:      "c++",
	CCompiler:        "c",
	ASMCompiler:      "as",
	ResourceCompiler: "rc",
	Archiver:         "ar",
	DLLLinker:        "dll",
	EXELinker:        "exe",
	OBJCopy:          "objcopy",
	OBJDump:          "objdump",
	Runlib:           "runlib",
	Strip:            "strip",
}

var longToolNames = map[Tool]string{
	UnknownTool:      "UNKNOWN_TOOL",
	CXXCompiler:      "C++ Compiler",
	CCompiler:        "C Compiler",
	ASMCompiler:      "Assembler Compiler",
	ResourceCompiler: "Resource Compiler",
	Archiver:         "Archiver",
	DLLLinker:        "DLL/SO Linker",
	EXELinker:        "Executable Linker",
	OBJCopy:          "objcopy",
	OBJDump:          "objdump",
	Runlib:           "runlib",
	Strip:            "strip",
}

// String implements Stringer interace for Tool
func (t Tool) String() string {
	ret := shortToolNames[t]
	if ret == "" {
		ret = "INVALID_TOOL"
	}
	return ret
}

// LongName returns long descriptive tool name
func (t Tool) LongName() string {
	ret := longToolNames[t]
	if ret == "" {
		ret = "INVALID_TOOL"
	}
	return ret
}

// ToolFromString converts text to Tool
func ToolFromString(s string) (Tool, error) {
	switch s {
	case "c++":
		return CXXCompiler, nil
	case "cpp":
		return CXXCompiler, nil
	case "cxx":
		return CXXCompiler, nil
	case "c":
		return CCompiler, nil
	case "cc":
		return CCompiler, nil
	case "as":
		return ASMCompiler, nil
	case "rc":
		return ResourceCompiler, nil
	case "ar":
		return Archiver, nil
	case "lib":
		return Archiver, nil
	case "dll":
		return DLLLinker, nil
	case "so":
		return DLLLinker, nil
	case "exe":
		return EXELinker, nil
	case "objcopy":
		return OBJCopy, nil
	case "objdump":
		return OBJDump, nil
	case "runlib":
		return Runlib, nil
	case "strip":
		return Strip, nil
	case "":
		return UnknownTool, errors.New("empty tool specifier")
	default:
		return UnknownTool, fmt.Errorf("unknown tool specifier '%s'", s)
	}
}

// MarshalText implements TextMarshaler interface for Tool
func (t *Tool) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements TextUnmarshaler interface for Tool
func (t *Tool) UnmarshalText(text []byte) (err error) {
	*t, err = ToolFromString(string(text))
	return
}

// MarshalJSON provides JSON writing support for Tool
func (t *Tool) MarshalJSON() (text []byte, err error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON provides JSON reading support for tool
func (t *Tool) UnmarshalJSON(text []byte) (err error) {
	*t, err = ToolFromString(string(text))
	return
}

// MarshalJSON provides JSON writing support for Toolset
func (t *Toolset) MarshalJSON() (text []byte, err error) {
	lines := []string{}
	for _, tool := range orderedToolList {
		if tool == UnknownTool {
			continue
		}
		fn, found := (*t)[tool]
		if !found {
			continue
		}
		fp, _ := json.Marshal(fn)
		lines = append(lines, fmt.Sprintf(`"%s":%s`, tool, string(fp)))
	}

	if len(lines) == 0 {
		return []byte("{}"), nil
	}
	s := "{" + strings.Join(lines, ",") + "}"
	return []byte(s), nil
}

var orderedToolList = []Tool{
	UnknownTool,
	CXXCompiler,
	CCompiler,
	ASMCompiler,
	ResourceCompiler,
	Archiver,
	DLLLinker,
	EXELinker,
	OBJCopy,
	OBJDump,
	Runlib,
	Strip,
}
