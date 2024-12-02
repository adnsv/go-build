package clang

import (
	"errors"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adnsv/go-build/compiler/gcc"
	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-build/compiler/triplet"
)

// Implementation identifies the specific LLVM-based compiler variant
type Implementation string

const (
	Clang      Implementation = "clang"
	AppleClang Implementation = "apple-clang"
	EmScripten Implementation = "emscripten"
	IntelClang Implementation = "intel-clang"
	TIClang    Implementation = "ti-clang"
	ARMClang   Implementation = "arm-clang"
	ZigClang   Implementation = "zig-clang"
)

// Ver contains version information extracted from compiler output
type Ver struct {
	Implementation Implementation `json:"implementation" yaml:"implementation"` // Specific compiler variant
	FullVersion    string         `json:"full-version" yaml:"full-version"`
	Version        string         `json:"version" yaml:"version"`
	Target         triplet.Full   `json:"target" yaml:"target"` // primary target extracted from compiler's version output
	ThreadModel    string         `json:"thread-model" yaml:"thread-model"`
	InstalledDir   string         `json:"installed-dir" yaml:"installed-dir"`
	CCIncludeDirs  []string       `json:"cc-include-dirs" yaml:"cc-include-dirs"`
	CXXIncludeDirs []string       `json:"cxx-include-dirs" yaml:"cxx-include-dirs"`
}

// Version detection regexes for different implementations
var (
	reClangVersion      = regexp.MustCompile(`^(?:.*clang) version ([\d\.]+)`)
	reAppleVersion      = regexp.MustCompile(`^Apple (?:clang|LLVM) version ([\d\.]+)`)
	reEmscriptenVersion = regexp.MustCompile(`^emcc \(Emscripten gcc/clang-like replacement.*\) ([\d\.]+)`)
	reIntelVersion      = regexp.MustCompile(`^Intel[^\n]+oneAPI[^\n]+ ([\d\.]+)`)
	reTIVersion         = regexp.MustCompile(`^TI .* Clang ([\d\.]+)`)
	reARMVersion        = regexp.MustCompile(`^armclang version ([\d\.]+)`)
	reZigVersion        = regexp.MustCompile(`^(?:Homebrew )?clang version ([\d\.]+)`)
)

var (
	reTarget       = regexp.MustCompile(`Target:\s+(.*)`)
	reThreadModel  = regexp.MustCompile(`Thread model:\s+(.*)`)
	reInstalledDir = regexp.MustCompile(`InstalledDir:\s+(.*)`)
)

// QueryVersionWithRegex is a generic version query function that uses a specific regex
func QueryVersionWithRegex(tool toolchain.ToolPath, impl Implementation, versionRegex *regexp.Regexp) (*Ver, error) {
	args := append(tool.Commands(), "-v")
	cmd := exec.Command(tool.Path(), args...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := string(buf)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return nil, errors.New("invalid version output")
	}

	match := versionRegex.FindStringSubmatch(strings.TrimSpace(lines[0]))
	if len(match) != 2 {
		return nil, errors.New("invalid version output")
	}

	v := match[1]
	if i := strings.IndexByte(v, '-'); i >= 0 {
		v = v[:i]
	}

	ret := &Ver{
		Implementation: impl,
		FullVersion:    strings.TrimSpace(match[0]),
		Version:        v,
	}

	// Special handling for emscripten
	if impl == EmScripten {
		ret.Target, _ = triplet.ParseFull("wasm32-emscripten")
	} else {
		for _, line := range lines {
			n := len(line)
			if n == 0 {
				continue
			}
			if line[n-1] == '\r' {
				line = line[:n-1]
			}
			match = reTarget.FindStringSubmatch(line)
			if len(match) == 2 {
				var err error
				ret.Target, err = triplet.ParseFull(strings.TrimSpace(match[1]))
				if err != nil {
					ret.Target = triplet.Full{Original: strings.TrimSpace(match[1])}
				}
			}
		}
	}

	for _, line := range lines {
		n := len(line)
		if n == 0 {
			continue
		}
		if line[n-1] == '\r' {
			line = line[:n-1]
		}
		match = reThreadModel.FindStringSubmatch(line)
		if len(match) == 2 {
			ret.ThreadModel = strings.TrimSpace(match[1])
		}
		match = reInstalledDir.FindStringSubmatch(output)
		if len(match) == 2 {
			ret.InstalledDir = filepath.ToSlash(strings.TrimSpace(match[1]))
		}
	}
	if includes, err := gcc.GetSystemIncludes(string(tool), "c"); err == nil {
		ret.CCIncludeDirs = append(ret.CCIncludeDirs, includes...)
	}
	if includes, err := gcc.GetSystemIncludes(string(tool), "c++"); err == nil {
		ret.CXXIncludeDirs = append(ret.CXXIncludeDirs, includes...)
	}
	return ret, nil
}

// QueryVersion attempts to detect the implementation and query its version
func QueryVersion(tool toolchain.ToolPath) (*Ver, error) {
	args := append(tool.Commands(), "-v")
	cmd := exec.Command(tool.Path(), args...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := strings.TrimSpace(strings.Split(string(buf), "\n")[0])

	// Try each implementation in order
	switch {
	case reEmscriptenVersion.MatchString(output):
		return QueryVersionWithRegex(tool, EmScripten, reEmscriptenVersion)
	case reAppleVersion.MatchString(output):
		return QueryVersionWithRegex(tool, AppleClang, reAppleVersion)
	case reIntelVersion.MatchString(output):
		return QueryVersionWithRegex(tool, IntelClang, reIntelVersion)
	case reTIVersion.MatchString(output):
		return QueryVersionWithRegex(tool, TIClang, reTIVersion)
	case reARMVersion.MatchString(output):
		return QueryVersionWithRegex(tool, ARMClang, reARMVersion)
	case reZigVersion.MatchString(output):
		return QueryVersionWithRegex(tool, ZigClang, reZigVersion)
	case reClangVersion.MatchString(output):
		return QueryVersionWithRegex(tool, Clang, reClangVersion)
	default:
		return nil, errors.New("unknown clang implementation")
	}
}
