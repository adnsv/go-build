package triplet

import (
	"strings"
)

type Target struct {
	OS   string `json:"os,omitempty" yaml:"os,omitempty"`
	Arch string `json:"arch,omitempty" yaml:"arch,omitempty"`
	ABI  string `json:"abi,omitempty" yaml:"abi,omitempty"`
	LibC string `json:"libc,omitempty" yaml:"libc,omitempty"`
}

type Full struct {
	Target
	Original string   `json:"original,omitempty" yaml:"original,omitempty"`
	Vendors  []string `json:"vendors,omitempty" yaml:"vendors,omitempty"`
}

func ParseFull(target string) Full {
	f := Full{Original: target}
	f.Target, f.Vendors = ParseTarget(target)
	return f
}

func ParseTarget(target string) (t Target, vendors []string) {
	t = Target{
		OS:   "unknown",
		Arch: "unknown",
		ABI:  "unknown",
		LibC: "unknown",
	}

	segments := strings.Split(target, "-")
	skip := map[string]struct{}{}

	// find Arch
	for _, s := range segments {
		if v, ok := ParseArch(s); ok {
			t.Arch = v
			skip[s] = struct{}{}
			break
		}
	}

	// find OS
	for _, s := range segments {
		if v, ok := ParseOS(s); ok {
			t.OS = v
			skip[s] = struct{}{}
			break
		}
	}

	// find ABI
	for _, s := range segments {
		if v, ok := ParseABI(s); ok {
			t.ABI = v
			skip[s] = struct{}{}
			break
		}
	}

	// find LibC
	for _, s := range segments {
		if v, ok := ParseLibC(s); ok {
			t.LibC = v
			skip[s] = struct{}{}
			break
		}
	}

	for _, s := range segments {
		if _, other := skip[s]; other {
			continue
		} else {
			vendors = append(vendors, s)
		}
	}
	return
}

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

func ParseArch(arch string) (string, bool) {
	var archNorm = map[string]string{
		"x64":       "x64",
		"amd64":     "x64",
		"x86_64":    "x64",
		"x32":       "x32",
		"86":        "x32",
		"x86":       "x32",
		"386":       "x32",
		"i386":      "x32",
		"486":       "x32",
		"i486":      "x32",
		"586":       "x32",
		"i586":      "x32",
		"686":       "x32",
		"i686":      "x32",
		"arm":       "arm",
		"arm32":     "arm",
		"arm64":     "arm64",
		"aarch64":   "arm64",
		"ia64":      "ia64",
		"powerpc":   "powerpc",
		"powerpcle": "powerpcle",
		"s390":      "s390",
		"s390x":     "s390x",
		"sparc":     "sparc",
		"sparc64":   "sparc64",
		"sparcv9":   "sparc64",
		"c6x":       "c6x",
		"tilegx":    "tilegx",
		"tilegxbe":  "tilegxbe",
		"tilepro":   "tilepro",
	}
	arch = strings.ToLower(arch)
	if norm, ok := archNorm[arch]; ok {
		arch = norm
	}
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

func NormalizeArch(arch string) string {
	n, _ := ParseArch(arch)
	return n
}

func ParseOS(os string) (string, bool) {
	os = strings.ToLower(os)
	switch os {
	case "mingw32", "mingw", "mingw64", "w64", "msvc", "windows":
		return "windows", true
	case "none", "cygwin", "msys",
		"freebsd", "netbsd", "openbsd",
		"vxworks", "vxworksae", "haiku":
		return os, true
	default:
		for _, p := range []string{"linux", "solaris", "darwin", "uclinux"} {
			if strings.HasPrefix(os, p) {
				return p, true
			}
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
	switch abi {
	case "eabi":
		return "eabi", true
	case "eabisim":
		return "eabisim", true
	case "mingw32", "mingw", "mingw64", "w64", "msvc", "windows", "cygwin", "msys":
		return "pe", true
	case "elf", "netbsd", "openbsd", "freebsd", "aix", "gnueabi", "gnueabihf":
		return "elf", true
	default:
		if strings.HasPrefix(abi, "linux") || strings.HasPrefix(abi, "uclinux") || strings.HasPrefix(abi, "solaris") {
			return "elf", true
		} else if strings.HasPrefix(abi, "darwin") {
			return "marcho", true
		} else {
			return abi, false
		}
	}
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
