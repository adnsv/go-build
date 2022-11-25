package msvc

import (
	"fmt"
	"io"
)

// Installation corresponds to an instance of VisualStudio
type Installation struct {
	DisplayName         string `json:"display-name" yaml:"display-name"`
	InstanceID          string `json:"instance-id" yaml:"instance-id"`
	InstallationPath    string `json:"installation-path" yaml:"installation-path"`
	InstallationVersion string `json:"installation-version" yaml:"installation-version"`
	Description         string `json:"description" yaml:"description"`
	IsPrerelease        bool   `json:"is-prerelease" yaml:"is-prerelease"`
}

func (i *Installation) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "%s\n", i.DisplayName)
	fmt.Fprintf(w, "- version: '%s'\n", i.InstallationVersion)
	fmt.Fprintf(w, "- instance id: '%s'\n", i.InstanceID)
	fmt.Fprintf(w, "- path: '%s'\n", i.InstallationPath)
	if i.IsPrerelease {
		fmt.Fprintf(w, "- this is a pre-release build")
	}
}
