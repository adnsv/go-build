package proj

import (
	"path/filepath"

	"github.com/adnsv/go-build/compiler/toolchain"
)

type Project struct {
	BuildRoot string
	Targets   []*Target
}

type Target struct {
	Name     string
	Features []*Feature
}

type Feature struct {
	Buildable
}

type Buildable struct {
	Name        string
	Includes    []string
	Sources     []string
	Definitions []string
	Libraries   map[string]LibraryRef
}

type LibraryRef interface {
}

type LibTarget struct {
	Target
}

type ExeTarget struct {
	Target
}

type SystemLibrary string

type ExternalLibrary struct {
}

func (t *Target) Enabled() bool {
	return true
}

func (f *Feature) Enabled() bool {
	return true
}

func (f Feature) RequiredTools() map[toolchain.Tool]bool {
	ret := map[toolchain.Tool]bool{}
	for _, s := range f.Sources {
		t := ExtensionTools[GetExtensionType(filepath.Ext(s))]
		if t != toolchain.UnknownTool {
			ret[t] = true
		}
	}
	return ret
}

func (t Target) RequiredTools() map[toolchain.Tool]bool {
	ret := map[toolchain.Tool]bool{}
	for _, f := range t.Features {
		if f.Enabled() {
			for tool := range f.RequiredTools() {
				ret[tool] = true
			}
		}
	}
	return ret
}

func (p Project) RequiredTools() map[toolchain.Tool]bool {
	ret := map[toolchain.Tool]bool{}
	for _, t := range p.Targets {
		if t.Enabled() {
			for tool := range t.RequiredTools() {
				ret[tool] = true
			}
		}
	}
	return ret
}

func (p Project) Buildables() []*Buildable {
	ret := []*Buildable{}
	for _, t := range p.Targets {
		if t.Enabled() {
			for _, f := range t.Features {
				if f.Enabled() {
					ret = append(ret, &f.Buildable)
				}
			}
		}
	}
	return ret
}
