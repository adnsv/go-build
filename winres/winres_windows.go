package winres

import (
	"github.com/adnsv/go-utils/version"
	"github.com/josephspurrier/goversioninfo"
	"github.com/winlabs/gowin32"
)

func UpdateVersionInfo(filename string, ver *goversioninfo.VersionInfo) error {
	ru, err := gowin32.NewResourceUpdate(filename, false)
	if err != nil {
		return err
	}
	ver.Build()
	ver.Walk()
	err = ru.Update(gowin32.ResourceTypeVersion, gowin32.IntResourceId(1), gowin32.Language(0x0409), ver.Buffer.Bytes())
	if err != nil {
		return err
	}
	return ru.Save()
}
