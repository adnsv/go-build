package main

import (
	"runtime/debug"
)

var app_ver string = ""

func app_version() string {
	v, ok := debug.ReadBuildInfo()
	if ok && v.Main.Version != "(devel)" {
		// installed with go install
		return v.Main.Version
	} else if app_ver != "" {
		// built with ld-flags
		return app_ver
	} else {
		return "#UNAVAILABLE"
	}
}
