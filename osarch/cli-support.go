package osarch

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

// ParseList parses one or more strings with comma-separated list of arches into
// a set of pairs, also supports some predefined values. All targets are
// normalized and only unique entries are returned.
func ParseList(ss ...string) ([]Pair, error) {
	targets := []Pair{}

	// make sure that the returned stuff is unique
	add := func(p Pair) {
		p = p.Normalized()
		if !slices.Contains(targets, p) {
			targets = append(targets, p)
		}
	}

	for _, s := range ss {
		for _, it := range strings.Split(s, ",") {
			t := strings.TrimSpace(it)
			switch t {
			case "":
				continue
			case "this", "host":
				add(This())
			case "windows":
				add(MustParse("windows-386"))
				add(MustParse("windows-amd64"))
				add(MustParse("windows-arm64"))
			case "common":
				add(MustParse("windows-386"))
				add(MustParse("windows-amd64"))
				add(MustParse("windows-arm64"))
				add(MustParse("linux-386"))
				add(MustParse("linux-amd64"))
				add(MustParse("linux-arm"))
				add(MustParse("linux-arm64"))
				add(MustParse("darwin-amd64"))
				add(MustParse("darwin-arm64"))
				add(MustParse("freebsd-amd64"))
				add(MustParse("freebsd-arm64"))
			default:
				pa, err := Parse(t)
				if err != nil {
					return nil, fmt.Errorf("failed to parse '%s': %w", t, err)
				}
				add(pa)
			}
		}
	}

	slices.SortFunc(targets, func(a, b Pair) bool {
		if a.OS < b.OS {
			return true
		}
		if a.OS > b.OS {
			return false
		}
		return a.Arch < b.Arch
	})

	return targets, nil
}
