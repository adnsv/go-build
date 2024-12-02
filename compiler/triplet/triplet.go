package triplet

import (
	"fmt"
	"strings"
)

// Package triplet provides functionality for parsing and handling compiler target triplets.
// Target triplets are commonly used in compiler toolchains to specify the target architecture,
// operating system, and ABI (e.g., x86_64-linux-gnu, aarch64-linux-android).

// archNorm maps various architecture names to their normalized form.
// This includes common variants and aliases for CPU architectures.
var archNorm = map[string]string{
	"x64":       "x64",       // 64-bit x86
	"amd64":     "x64",       // AMD64/Intel64
	"x86_64":    "x64",       // Standard x86_64
	"x32":       "x32",       // x32 ABI
	"86":        "x32",       // Generic x86
	"x86":       "x32",       // 32-bit x86
	"386":       "x32",       // Intel 386
	"i386":      "x32",       // Intel 386
	"486":       "x32",       // Intel 486
	"i486":      "x32",       // Intel 486
	"586":       "x32",       // Intel Pentium
	"i586":      "x32",       // Intel Pentium
	"686":       "x32",       // Intel Pentium Pro
	"i686":      "x32",       // Intel Pentium Pro
	"arm":       "arm",       // 32-bit ARM
	"arm32":     "arm",       // 32-bit ARM
	"arm64":     "arm64",     // 64-bit ARM
	"aarch64":   "arm64",     // ARM64
	"ia64":      "ia64",      // Intel Itanium
	"powerpc":   "powerpc",   // PowerPC
	"powerpcle": "powerpcle", // PowerPC Little Endian
	"s390":      "s390",      // IBM System/390
	"s390x":     "s390x",     // IBM System/390x
	"sparc":     "sparc",     // SPARC
	"sparc64":   "sparc64",   // SPARC64
	"sparcv9":   "sparc64",   // SPARC V9
	"c6x":       "c6x",       // TI C6x DSP
	"tilegx":    "tilegx",    // Tilera TILE-Gx
	"tilegxbe":  "tilegxbe",  // Tilera TILE-Gx Big Endian
	"tilepro":   "tilepro",   // Tilera TILEPro
}

// abiMap maps various ABI/environment names to their normalized form.
// This includes common toolchain environments and binary interfaces.
var abiMap = map[string]string{
	"eabi":      "eabi",    // Embedded ABI
	"eabisim":   "eabisim", // Embedded ABI Simulator
	"mingw32":   "pe",      // MinGW 32-bit
	"mingw":     "pe",      // MinGW
	"mingw64":   "pe",      // MinGW 64-bit
	"w64":       "pe",      // Windows 64-bit
	"msvc":      "pe",      // Microsoft Visual C++
	"windows":   "pe",      // Windows
	"cygwin":    "pe",      // Cygwin
	"msys":      "pe",      // MSYS
	"elf":       "elf",     // ELF format
	"netbsd":    "elf",     // NetBSD
	"openbsd":   "elf",     // OpenBSD
	"freebsd":   "elf",     // FreeBSD
	"aix":       "elf",     // AIX
	"gnueabi":   "elf",     // GNU EABI
	"gnueabihf": "elf",     // GNU EABI Hard Float
}

// Target represents a normalized compiler target triplet.
// The format typically follows: architecture-operating_system-environment
type Target struct {
	OS   string `json:"os,omitempty" yaml:"os,omitempty"`     // Operating system (e.g., linux, windows)
	Arch string `json:"arch,omitempty" yaml:"arch,omitempty"` // Architecture (e.g., x64, arm64)
	ABI  string `json:"abi,omitempty" yaml:"abi,omitempty"`   // ABI/environment (e.g., elf, pe)
	LibC string `json:"libc,omitempty" yaml:"libc,omitempty"` // C library (e.g., glibc, msvcrt)
}

// Full extends Target with original string and vendor information
type Full struct {
	Target   `yaml:",inline"`
	Original string   `json:"original,omitempty" yaml:"original,omitempty"`    // Original triplet string
	Vendors  []string `json:"vendors,omitempty" yaml:"vendors,flow,omitempty"` // Vendor-specific components
}

// ParseFull parses a target triplet string into a Full struct
func ParseFull(target string) (Full, error) {
	f := Full{Original: target}
	t, v, err := ParseTarget(target)
	if err != nil {
		return Full{}, err
	}
	f.Target = t
	f.Vendors = v
	return f, nil
}

// ParseTarget parses a target triplet string into its components.
// It attempts to identify the architecture, OS, ABI, and C library
// from the hyphen-separated components of the target string.
func ParseTarget(target string) (t Target, vendors []string, err error) {
	if target == "" {
		return Target{}, []string{}, fmt.Errorf("empty target string")
	}

	t = Target{
		OS:   "unknown",
		Arch: "unknown",
		ABI:  "unknown",
		LibC: "unknown",
	}
	vendors = []string{}

	segments := strings.Split(target, "-")
	skip := map[string]struct{}{}

	// Find architecture component
	for _, s := range segments {
		if v, ok := ParseArch(s); ok {
			if t.Arch == "unknown" {
				t.Arch = v
				skip[s] = struct{}{}
			}
			break
		}
	}

	// Find operating system component
	for _, s := range segments {
		if v, ok := ParseOS(s); ok {
			if t.OS == "none" || t.OS == "unknown" {
				t.OS = v
				skip[s] = struct{}{}
			}
		}
	}

	// Find ABI component
	for _, s := range segments {
		if v, ok := ParseABI(s); ok {
			if t.ABI == "unknown" {
				t.ABI = v
				skip[s] = struct{}{}
			}
		}
	}

	// Collect vendor components
	for _, s := range segments {
		if _, skipped := skip[s]; !skipped {
			vendors = append(vendors, s)
		}
	}

	// Attempt to identify C library from vendor components
	for i, s := range vendors {
		if v, ok := ParseLibC(s); ok {
			t.LibC = v
			if s == "msvc" || s == "mingw" || s == "mingw32" || s == "mingw64" || s == "w64" {
				vendors = append(vendors[:i], vendors[i+1:]...)
			}
			break
		}
	}

	return
}

// Match checks if this target matches another target.
// Empty fields in either target are treated as wildcards.
func (t *Target) Match(other Target) bool {
	if t.Arch != "" && other.Arch != "" && t.Arch != other.Arch {
		return false
	}
	if t.OS != "" && other.OS != "" && t.OS != other.OS {
		return false
	}
	if t.ABI != "" && other.ABI != "" && t.ABI != other.ABI {
		return false
	}
	if t.LibC != "" && other.LibC != "" && t.LibC != other.LibC {
		return false
	}
	return true
}

// ParseArch parses and normalizes an architecture string.
// Returns the normalized architecture name and whether it was recognized.
func ParseArch(arch string) (string, bool) {
	arch = strings.ToLower(arch)
	if norm, ok := archNorm[arch]; ok {
		return norm, true
	}
	// Handle architecture prefixes for less common architectures
	for _, p := range []string{
		"aarch64", "amdgcn", "arc", "arm", "avr", "blackfin", "cr16",
		"cris", "epiphany", "h8300", "ia64", "iq2000", "lm32", "m32c", "m32r",
		"m68k", "microblaze", "mips", "moxie", "msp430", "nds32le", "nds32be",
		"nvptx", "or1k", "rl78", "riscv32", "riscv64", "rx", "xtensa", "visium",
	} {
		if strings.HasPrefix(arch, p) {
			return arch, true
		}
	}
	return arch, false
}

// NormalizeArch normalizes an architecture string to its standard form
func NormalizeArch(arch string) string {
	n, _ := ParseArch(arch)
	return n
}

// osMap maps various OS names to their normalized form.
// This includes common operating systems and their variants.
var osMap = map[string]string{
	// Windows variants
	"mingw32": "windows",
	"mingw":   "windows",
	"mingw64": "windows",
	"w64":     "windows",
	"msvc":    "windows",
	"windows": "windows",
	// Unix-like systems
	"linux":     "linux",
	"darwin":    "darwin",
	"freebsd":   "freebsd",
	"netbsd":    "netbsd",
	"openbsd":   "openbsd",
	"dragonfly": "dragonfly",
	"solaris":   "solaris",
	"sunos":     "solaris",
	"illumos":   "solaris",
	"aix":       "aix",
	"hpux":      "hpux",
	"ios":       "ios",
	// Embedded/Special
	"uclinux":    "uclinux",
	"none":       "none",
	"baremetal":  "none",
	"cygwin":     "cygwin",
	"msys":       "msys",
	"vxworks":    "vxworks",
	"vxworksae":  "vxworks",
	"haiku":      "haiku",
	"android":    "android",
	"emscripten": "emscripten",
	"wasi":       "wasi",
}

// ParseOS parses and normalizes an operating system string.
// Returns the normalized OS name and whether it was recognized.
func ParseOS(os string) (string, bool) {
	os = strings.ToLower(os)
	if normalized, ok := osMap[os]; ok {
		return normalized, true
	}
	// Handle prefixed versions (e.g., linux-gnu, darwin20)
	for prefix, normalized := range osMap {
		if strings.HasPrefix(os, prefix) {
			return normalized, true
		}
	}
	return os, false
}

func NormalizeOS(os string) string {
	n, _ := ParseOS(os)
	return n
}

func ParseABI(abi string) (string, bool) {
	abi = strings.ToLower(abi)
	if normalized, ok := abiMap[abi]; ok {
		return normalized, true
	}

	// Handle prefixes
	if strings.HasPrefix(abi, "linux") ||
		strings.HasPrefix(abi, "uclinux") ||
		strings.HasPrefix(abi, "solaris") {
		return "elf", true
	}
	if strings.HasPrefix(abi, "darwin") {
		return "marcho", true
	}

	return abi, false
}

func NormalizeABI(abi string) string {
	n, _ := ParseABI(abi)
	return n
}

func ParseLibC(libc string) (string, bool) {
	libc = strings.ToLower(libc)
	switch libc {
	case "mingw32", "mingw", "mingw64", "w64":
		return "mingw", true
	case "musl":
		return "musl", true
	case "gnu", "msys", "cygwin", "glibc":
		return "glibc", true
	case "mcvcrt", "msvc":
		return "msvcrt", true
	default:
		return libc, false
	}
}

func NormalizeLibC(libc string) string {
	n, _ := ParseLibC(libc)
	return n
}

func (t Target) IsValid() bool {
	return t.OS != "unknown" && t.Arch != "unknown"
}

func (t Target) String() string {
	var parts []string
	if t.Arch != "" && t.Arch != "unknown" {
		parts = append(parts, t.Arch)
	}
	if t.OS != "" && t.OS != "unknown" {
		parts = append(parts, t.OS)
	}
	if t.ABI != "" && t.ABI != "unknown" {
		parts = append(parts, t.ABI)
	}
	if t.LibC != "" && t.LibC != "unknown" {
		parts = append(parts, t.LibC)
	}
	return strings.Join(parts, "-")
}

// Define specific error types
type ErrInvalidTarget struct {
	Target string
	Reason string
}

func (e ErrInvalidTarget) Error() string {
	return fmt.Sprintf("invalid target %q: %s", e.Target, e.Reason)
}

// Add validation method
func (t Target) Validate() error {
	if t.Arch == "unknown" {
		return &ErrInvalidTarget{t.String(), "unknown architecture"}
	}
	if t.OS == "unknown" {
		return &ErrInvalidTarget{t.String(), "unknown operating system"}
	}
	return nil
}

// Add constructor for Target
func NewTarget(arch, os, abi, libc string) Target {
	return Target{
		Arch: NormalizeArch(arch),
		OS:   NormalizeOS(os),
		ABI:  NormalizeABI(abi),
		LibC: NormalizeLibC(libc),
	}
}

// Add helper methods for common checks
func (t Target) IsDarwin() bool {
	return t.OS == "darwin" || t.OS == "ios"
}

func (t Target) IsLinux() bool {
	return t.OS == "linux" || t.OS == "android" || t.OS == "uclinux"
}

func (t Target) IsBSD() bool {
	return t.OS == "freebsd" || t.OS == "netbsd" || t.OS == "openbsd" || t.OS == "dragonfly"
}

func (t Target) IsSolaris() bool {
	return t.OS == "solaris" || t.OS == "sunos" || t.OS == "illumos"
}

func (t Target) IsUnix() bool {
	return t.IsLinux() || t.IsDarwin() || t.IsBSD() || t.IsSolaris() ||
		t.OS == "aix" || t.OS == "hpux"
}

func (t Target) IsPOSIX() bool {
	return t.IsUnix() || t.OS == "cygwin" || t.OS == "msys"
}

func (t Target) IsEmbedded() bool {
	return t.OS == "none" || t.OS == "baremetal" || t.OS == "vxworks"
}

func (t Target) IsWasm() bool {
	return t.OS == "emscripten" || t.OS == "wasi"
}
