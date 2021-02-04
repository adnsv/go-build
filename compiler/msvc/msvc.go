package msvc

import (
	"fmt"
	"io"
)

// Installation corresponds to an instance of VisualStudio
type Installation struct {
	InstanceID          string `json:"instanceId"`
	DisplayName         string `json:"displayName"`
	InstallationPath    string `json:"installationPath"`
	InstallationVersion string `json:"installationVersion"`
	Description         string `json:"description"`
	IsPrerelease        bool   `json:"isPrerelease"`
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
