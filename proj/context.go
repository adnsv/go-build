package proj

import (
	"github.com/adnsv/go-build/compiler/toolchain"
)

type Context struct {
	Toolchain *toolchain.Toolset
	Flagset   Flagset
}

type Flagset map[toolchain.Tool]Flags

type Flags []string

func (ff Flags) Contains(flg string) bool {
	for _, f := range ff {
		if f == flg {
			return true
		}
	}
	return false
}

func removeDuplicates(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	ret := make([]string, 0, len(ss))
	for _, s := range ss {
		if s == "" {
			continue
		}
		if _, exists := seen[s]; !exists {
			seen[s] = struct{}{}
			ret = append(ret, s)
		}
	}
	return ret
}
