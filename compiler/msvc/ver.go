package msvc

import (
	"errors"
	"fmt"
)

type VersionQuad struct {
	Major int
	Minor int
	Patch int
	Build int
}

var ErrInvalidVersionQuad = errors.New("invalid version quad")

func ParseVersionQuad(s string) (v VersionQuad, err error) {
	n, err := fmt.Sscanf(s, "%d.%d.%d.%d", &v.Major, &v.Major, &v.Patch, &v.Build)
	if n != 4 {
		err = ErrInvalidVersionQuad
	}
	return
}

func (v VersionQuad) Compare(o VersionQuad) int {
	if v.Major != o.Major {
		return v.Major - o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor - o.Minor
	}
	if v.Patch != o.Patch {
		return v.Patch - o.Patch
	}
	return v.Build - o.Build
}
