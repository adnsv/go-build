package osarch

import (
	"encoding/json"
	"errors"
	"runtime"
	"strings"
)

// Pair contains a OS, architecture pair, typically pulled from runtime.GOOS and runtime.GOARCH
type Pair struct {
	OS   string
	Arch string
}

// This returns runtime.GOOS, runtime.GOARCH pair,
// This is the pair for which the compiler is currently building
func This() Pair {
	return Pair{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// soqm is a pass through for a string that returns question marks for empty inputs.
func soqm(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

// String implements Stringer interface for Pair
func (oa Pair) String() string {
	return soqm(oa.OS) + "-" + soqm(oa.Arch)
}

// NormArch maps alternative arch representations to
// GO-compatible arch names
func NormArch(arch string) string {
	arch = strings.ToLower(arch)
	switch arch {
	case "x32":
		return "386"
	case "x64":
		return "amd64"
	case "x86_64":
		return "amd64"
	case "arm32":
		return "arm"
	default:
		return arch
	}
}

// AltArch maps arch values to an alternative representations
// - 386 -> x32
// - amd64 -> x64
// - arm -> arm32
// - arm64 -> arm64 (noop)
func AltArch(arch string) string {
	arch = strings.ToLower(arch)
	switch arch {
	case "386":
		return "x32"
	case "amd64":
		return "x64"
	case "arm":
		return "arm32"
	case "x86_64":
		return "amd64"
	default:
		return arch
	}
}

// Normalized converts Pair to its GO-normalized representation
func (oa Pair) Normalized() Pair {
	return Pair{OS: strings.ToLower(oa.OS), Arch: NormArch(oa.Arch)}
}

// Alted converts Pair to its alternative representation
func (oa Pair) Alted() Pair {
	return Pair{OS: strings.ToLower(oa.OS), Arch: AltArch(oa.Arch)}
}

// ExeExt returns default executable file extension for the target OS
func (oa Pair) ExeExt() string {
	if oa.OS == "windows" {
		return ".exe"
	}
	return ""
}

// SameArch checks if the two arch names refer to the same arch
func SameArch(a, b string) bool {
	return NormArch(a) == NormArch(b)
}

// MarshalJSON is used when writing Pair into JSON streams
func (oa *Pair) MarshalJSON() (bb []byte, err error) {
	bb, err = json.Marshal(oa.String())
	return
}

// MarshalText implements TextMarshaler interface for Pair
func (oa *Pair) MarshalText() (bb []byte, err error) {
	bb, err = []byte(oa.String()), nil
	return
}

// ErrInvalid is returned when a string representation of
// Pair is invalid
var ErrInvalid = errors.New("invalid os-arch")

// Parse parses a string into Pair
func Parse(s string) (Pair, error) {
	i := strings.IndexAny(s, "-._")
	if i < 1 || i >= len(s)-1 {
		return Pair{}, ErrInvalid
	}
	return Pair{OS: s[:i], Arch: s[i+1:]}, nil
}

// MustParse is like Parse, but panics isth string cannot be parsed
func MustParse(s string) Pair {
	ret, err := Parse(s)
	if err != nil {
		panic(`os-arch: Parse(` + s + `): ` + err.Error())
	}
	return ret
}

// UnmarshalJSON is used when reading Pair from JSON streams
func (oa *Pair) UnmarshalJSON(b []byte) (err error) {
	var s string
	if json.Unmarshal(b, &s); err == nil {
		*oa, err = Parse(s)
		return
	}
	return ErrInvalid
}

// UnmarshalText implements TextUnmarshaler interface for Pair
func (oa *Pair) UnmarshalText(text []byte) (err error) {
	*oa, err = Parse(string(text))
	return
}
