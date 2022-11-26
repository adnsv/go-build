package msvc

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
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
