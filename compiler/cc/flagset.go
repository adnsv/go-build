package cc

import (
	"fmt"
	"strings"
)

type FlagSet map[string]struct{}

func (fs *FlagSet) Add(flags ...string) {
	if (*fs) == nil {
		*fs = FlagSet{}
	}
	for _, flag := range flags {
		(*fs)[flag] = struct{}{}
	}
}

func (fs *FlagSet) Insert(other FlagSet) {
	if (*fs) == nil {
		*fs = FlagSet{}
	}
	for f := range other {
		(*fs)[f] = struct{}{}
	}
}

type BuildConfig int

const (
	All = BuildConfig(iota)
	Debug
	Release
	MinSizeRel
	RelWithDebInfo
)

type Flags map[BuildConfig]FlagSet

func (f *Flags) Add(c BuildConfig, flags ...string) {
	if len(flags) == 0 {
		return
	}
	if (*f) == nil {
		*f = map[BuildConfig]FlagSet{}
	}
	dst := (*f)[c]
	if dst == nil {
		dst = FlagSet{}
	}
	for _, flag := range flags {
		dst[flag] = struct{}{}
	}
	(*f)[c] = dst
}

func (f BuildConfig) String() string {
	switch f {
	case All:
		return "all"
	case Debug:
		return "debug"
	case Release:
		return "release"
	case MinSizeRel:
		return "minsizerel"
	case RelWithDebInfo:
		return "relwithdebinfo"
	default:
		return "#invalid"
	}
}

func FlagConfigFromString(s string) (BuildConfig, error) {
	switch strings.ToLower(s) {
	case "all":
		return All, nil
	case "debug":
		return Debug, nil
	case "release":
		return Release, nil
	case "minsizerel":
		return MinSizeRel, nil
	case "relwithdebinfo":
		return RelWithDebInfo, nil
	default:
		return All, fmt.Errorf("unknown flag config '%s'", s)
	}
}

func (f BuildConfig) MarshalText() (text []byte, err error) {
	return []byte(f.String()), nil
}

func (t *BuildConfig) UnmarshalText(text []byte) (err error) {
	*t, err = FlagConfigFromString(string(text))
	return
}
