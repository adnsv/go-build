// +build !windows

package msvc

import "github.com/adnsv/go-build/compiler/toolchain"

func DiscoverInstallations(feedback func(string)) ([]*Installation, error) {
	return nil, nil
}

func DiscoverToolchains(feedback func(string)) []*toolchain.Chain {
	return nil
}
