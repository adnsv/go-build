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
		if v.Major > o.Major {
			return 1
		}
		return -1
	}
	if v.Minor != o.Minor {
		if v.Minor > o.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != o.Patch {
		if v.Patch > o.Patch {
			return 1
		}
		return -1
	}
	if v.Build != o.Build {
		if v.Build > o.Build {
			return 1
		}
		return -1
	}
	return 0
}
