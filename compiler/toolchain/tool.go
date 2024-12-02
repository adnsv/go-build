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
	Linker
	ManifestTool

	OBJCopy
	OBJDump
	Ranlib
	Strip
)

// ToolPath represents a path to a tool executable with optional subcommands
// Format: path|cmd1|cmd2|... (pipe separator to avoid conflicts with Windows drive letters)
type ToolPath string

// NewToolPath creates a ToolPath from a path and optional subcommands
func NewToolPath(path string, commands ...string) ToolPath {
	if len(commands) == 0 {
		return ToolPath(path)
	}
	return ToolPath(path + "|" + strings.Join(commands, "|"))
}

// Path returns the executable path part (before any pipe)
func (tp ToolPath) Path() string {
	if i := strings.IndexByte(string(tp), '|'); i >= 0 {
		return string(tp[:i])
	}
	return string(tp)
}

// Commands returns the optional subcommands (after pipes) as a slice
func (tp ToolPath) Commands() []string {
	if i := strings.IndexByte(string(tp), '|'); i >= 0 {
		return strings.Split(string(tp[i+1:]), "|")
	}
	return nil
}

// HasCommands returns true if the tool path includes subcommands
func (tp ToolPath) HasCommands() bool {
	return strings.IndexByte(string(tp), '|') >= 0
}

// String returns the full tool path including subcommands
func (tp ToolPath) String() string {
	return string(tp)
}

// Toolset is a collection of tools with their paths and optional subcommands
type Toolset map[Tool]ToolPath

var shortToolNames = map[Tool]string{
	UnknownTool:      "UNKNOWN_TOOL",
	CXXCompiler:      "c++",
	CCompiler:        "c",
	ASMCompiler:      "as",
	ResourceCompiler: "rc",
	Archiver:         "ar",
	Linker:           "linker",
	ManifestTool:     "mt",
	OBJCopy:          "objcopy",
	OBJDump:          "objdump",
	Ranlib:           "ranlib",
	Strip:            "strip",
}

var longToolNames = map[Tool]string{
	UnknownTool:      "UNKNOWN_TOOL",
	CXXCompiler:      "C++ Compiler",
	CCompiler:        "C Compiler",
	ASMCompiler:      "Assembler Compiler",
	ResourceCompiler: "Resource Compiler",
	Archiver:         "Archiver",
	Linker:           "Linker",
	ManifestTool:     "Manifest Tool",
	OBJCopy:          "objcopy",
	OBJDump:          "objdump",
	Ranlib:           "ranlib",
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
	case "linker":
		return Linker, nil
	case "dll":
		return Linker, nil
	case "so":
		return Linker, nil
	case "exe":
		return Linker, nil
	case "mt":
		return ManifestTool, nil
	case "objcopy":
		return OBJCopy, nil
	case "objdump":
		return OBJDump, nil
	case "ranlib":
		return Ranlib, nil
	case "strip":
		return Strip, nil
	case "":
		return UnknownTool, errors.New("empty tool specifier")
	default:
		return UnknownTool, fmt.Errorf("unknown tool specifier '%s'", s)
	}
}

// MarshalText implements TextMarshaler interface for Tool
func (t Tool) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements TextUnmarshaler interface for Tool
func (t *Tool) UnmarshalText(text []byte) (err error) {
	*t, err = ToolFromString(string(text))
	return
}

func (t Tool) MarshalYAML() (interface{}, error) {
	return t.String(), nil
}

// MarshalJSON provides JSON writing support for Tool
func (t Tool) MarshalJSON() (text []byte, err error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON provides JSON reading support for tool
func (t *Tool) UnmarshalJSON(text []byte) (err error) {
	*t, err = ToolFromString(string(text))
	return
}

// Contains checks whether the specific tool exists
func (t Toolset) Contains(tool Tool) bool {
	_, ok := t[tool]
	return ok
}

// MarshalJSON provides JSON writing support for Toolset
func (t Toolset) MarshalJSON() (text []byte, err error) {
	lines := []string{}
	for _, tool := range orderedToolList {
		if tool == UnknownTool {
			continue
		}
		fn, found := t[tool]
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
	Linker,
	ManifestTool,
	OBJCopy,
	OBJDump,
	Ranlib,
	Strip,
}
