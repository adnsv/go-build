package winres

import (
	"github.com/adnsv/go-utils/version"
	"github.com/josephspurrier/goversioninfo"
	"github.com/winlabs/gowin32"
)

type VersionSet struct {
	Semantic version.Semantic
	Quad     version.Quad
}

func NewVersionSet(sem version.Semantic, additionalCommits int) (vs VersionSet, err error) {
	vs.Semantic = sem
	vs.Quad, err = version.MakeQuad(sem, additionalCommits)
	return vs, err
}

type Strings = goversioninfo.StringFileInfo

// NewVersionInfo creates and populates a versio info resource with some nice defaults
func NewVersionInfo(productver VersionSet, filever VersionSet, ss *Strings) *goversioninfo.VersionInfo {
	v := &goversioninfo.VersionInfo{}
	v.FileFlagsMask = "3F"
	v.FileFlags = "00"
	v.FileOS = "040004"
	v.FileType = "01"
	v.FileSubType = "00"
	v.VarFileInfo.Translation.LangID = goversioninfo.LangID(0x0409)
	v.VarFileInfo.Translation.CharsetID = goversioninfo.CharsetID(0x04B0)

	if ss != nil {
		v.StringFileInfo = *ss
	}

	v.FixedFileInfo.ProductVersion = goversioninfo.FileVersion(productver.Quad)
	v.StringFileInfo.ProductVersion = productver.Semantic.String()

	v.FixedFileInfo.FileVersion = goversioninfo.FileVersion(filever.Quad)
	v.StringFileInfo.FileVersion = filever.Semantic.String()

	return v
}

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
