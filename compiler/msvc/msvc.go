package msvc

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
)

// ArchitectureSpec represents a supported MSVC architecture configuration
type ArchitectureSpec struct {
	Name          string // Combined name (e.g., "amd64" or "amd64_x86")
	HostArch      string // Host architecture
	TargetArch    string // Target architecture
	CrossCompiler bool   // Whether this is a cross-compilation setup
}

// ToolsetVersion represents a specific MSVC toolset version
type ToolsetVersion struct {
	Version     string // Version string (e.g., "14.29.30133")
	Path        string // Path to the toolset
	IsDefault   bool   // Whether this is the default toolset
	SDKVersion  string // Windows SDK version compatible with this toolset
	UCRTVersion string // Universal CRT version
}

// Installation corresponds to an instance of Visual Studio
type Installation struct {
	DisplayName         string           `json:"display-name" yaml:"display-name"`
	InstanceID          string           `json:"instance-id" yaml:"instance-id"`
	InstallationPath    string           `json:"installation-path" yaml:"installation-path"`
	InstallationVersion string           `json:"installation-version" yaml:"installation-version"`
	Description         string           `json:"description" yaml:"description"`
	IsPrerelease        bool             `json:"is-prerelease" yaml:"is-prerelease"`
	ToolsetVersions     []ToolsetVersion `json:"toolset-versions" yaml:"toolset-versions"`
	DiscoveryMethod     string           `json:"discovery-method" yaml:"discovery-method"` // "vswhere", "env", "standalone"
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "%s\n", i.DisplayName)
	fmt.Fprintf(w, "- version: '%s'\n", i.InstallationVersion)
	fmt.Fprintf(w, "- instance id: '%s'\n", i.InstanceID)
	fmt.Fprintf(w, "- path: '%s'\n", i.InstallationPath)
	fmt.Fprintf(w, "- discovery method: '%s'\n", i.DiscoveryMethod)
	if len(i.ToolsetVersions) > 0 {
		fmt.Fprintf(w, "- toolsets:\n")
		for _, ts := range i.ToolsetVersions {
			fmt.Fprintf(w, "  - version: '%s'%s\n", ts.Version, map[bool]string{true: " (default)", false: ""}[ts.IsDefault])
			if ts.SDKVersion != "" {
				fmt.Fprintf(w, "    sdk: '%s'\n", ts.SDKVersion)
			}
			if ts.UCRTVersion != "" {
				fmt.Fprintf(w, "    ucrt: '%s'\n", ts.UCRTVersion)
			}
		}
	}
	if i.IsPrerelease {
		fmt.Fprintf(w, "- this is a pre-release build\n")
	}
}

// Extended environment variables used for MSVC detection
var msvcEnvVarsExtended = []string{
	"VS140COMNTOOLS", // VS 2015
	"VS120COMNTOOLS", // VS 2013
	"VS110COMNTOOLS", // VS 2012
	"VSINSTALLDIR",
	"VS_INSTALLDIR",
	"VS_PLATFORMTOOLSET",
	"CL",
	"_CL_",
	"INCLUDE",
	"LIBPATH",
	"LINK",
	"_LINK_",
	"LIB",
	"PATH",
	"TMP",
	"FRAMEWORKDIR",
	"FRAMEWORKDIR64",
	"FRAMEWORKVERSION",
	"FRAMEWORKVERSION64",
	"UCRTCONTEXTROOT",
	"UCRTVERSION",
	"UNIVERSALCRTSDKDIR",
	"VCINSTALLDIR",
	"VCTARGETSPATH",
	"WINDOWSLIBPATH",
	"WINDOWSSDKDIR",
	"WINDOWSSDKLIBVERSION",
	"WINDOWSSDKVERSION",
	"VISUALSTUDIOVERSION",
}

var reVersion = regexp.MustCompile("^Microsoft .*Version (.*) for (.*)")

func QueryVersion(exe string) (ver, target string, err error) {
	cmd := exec.Command(exe)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", err
	}
	match := reVersion.FindStringSubmatch(string(buf))
	if len(match) != 3 {
		return "", "", errors.New("unsupported version output")
	}
	return match[1], strings.TrimSpace(match[2]), nil
}

func Compare(c1, c2 *toolchain.Chain) int {
	q1, e1 := ParseVersionQuad(c1.Version)
	q2, e2 := ParseVersionQuad(c2.Version)
	if e1 == nil && e2 == nil {
		if i := q1.Compare(q2); i != 0 {
			return i
		}
	} else if e1 == nil {
		return -1
	} else if e2 == nil {
		return +1
	}
	q1, e1 = ParseVersionQuad(c1.WindowsSDKVersion)
	q2, e2 = ParseVersionQuad(c1.WindowsSDKVersion)
	if e1 == nil && e2 == nil {
		if i := q1.Compare(q2); i != 0 {
			return i
		}
	} else if e1 == nil {
		return -1
	} else if e2 == nil {
		return +1
	}
	q1, e1 = ParseVersionQuad(c1.UCRTVersion)
	q2, e2 = ParseVersionQuad(c1.UCRTVersion)
	if e1 == nil && e2 == nil {
		if i := q1.Compare(q2); i != 0 {
			return i
		}
	} else if e1 == nil {
		return -1
	} else if e2 == nil {
		return +1
	}
	if i := strings.Compare(c1.FullVersion, c2.FullVersion); i != 0 {
		return i
	}
	return strings.Compare(c1.InstalledDir, c2.InstalledDir)
}

var ToolNames = map[string]toolchain.Tool{
	"cl":   toolchain.CXXCompiler,
	"link": toolchain.Linker,
	"lib":  toolchain.Archiver,
	"rc":   toolchain.ResourceCompiler,
	"mt":   toolchain.ManifestTool,
}

var ToolEnvs = map[string]toolchain.Tool{
	"CL":   toolchain.CXXCompiler,
	"LINK": toolchain.Linker,
}
