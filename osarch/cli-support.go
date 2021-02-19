package osarch

import "strings"

// ParseList parses a string with comma-separated list of arches into a set of pairs,
// also supports some shortcuts.
func ParseList(s string) ([]Pair, error) {
	targets := []Pair{}
	if s == "this" || s == "host" {
		targets = append(targets, This())
	} else if s == "windows" {
		// build windows-x32 and windows-x64
		targets = append(targets,
			MustParse("windows-386"),
			MustParse("windows-amd64"),
		)
	} else if s == "common" {
		// common targets
		targets = append(targets,
			MustParse("windows-386"),
			MustParse("windows-amd64"),
			MustParse("linux-386"),
			MustParse("linux-amd64"),
			MustParse("linux-arm"),
			MustParse("linux-arm64"),
			MustParse("darwin-amd64"),
			MustParse("freebsd-amd64"),
		)
	} else {
		for _, t := range strings.Split(s, ",") {
			pa, err := Parse(strings.TrimSpace(t))
			if err != nil {
				return nil, err
			}
			targets = append(targets, pa)
		}
	}
	return targets, nil
}
