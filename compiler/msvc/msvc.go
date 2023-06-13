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

// Installation corresponds to an instance of VisualStudio
type Installation struct {
	DisplayName         string `json:"display-name" yaml:"display-name"`
	InstanceID          string `json:"instance-id" yaml:"instance-id"`
	InstallationPath    string `json:"installation-path" yaml:"installation-path"`
	InstallationVersion string `json:"installation-version" yaml:"installation-version"`
	Description         string `json:"description" yaml:"description"`
	IsPrerelease        bool   `json:"is-prerelease" yaml:"is-prerelease"`
	ToolsetVersion      string `json:"toolset-version" yaml:"toolset-version"`
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "%s\n", i.DisplayName)
	fmt.Fprintf(w, "- version: '%s'\n", i.InstallationVersion)
	fmt.Fprintf(w, "- instance id: '%s'\n", i.InstanceID)
	fmt.Fprintf(w, "- path: '%s'\n", i.InstallationPath)
	if i.ToolsetVersion != "" {
		fmt.Fprintf(w, "- toolset: '%s'\n", i.ToolsetVersion)
	}
	if i.IsPrerelease {
		fmt.Fprintf(w, "- this is a pre-release build")
	}
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
