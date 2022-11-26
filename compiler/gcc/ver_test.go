package gcc

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func Example_parseConfig() {
	const input = `Configured with: ../gcc-11.2.0/configure --prefix=/mingw64 --with-local-prefix=/mingw64/local --build=x86_64-w64-mingw32 --host=x86_64-w64-mingw32 --target=x86_64-w64-mingw32 --with-native-system-header-dir=/mingw64/include --libexecdir=/mingw64/lib --enable-bootstrap --enable-checking=release --with-arch=x86-64 --with-tune=generic --enable-languages=c,lto,c++,fortran,ada,objc,obj-c++,jit --enable-shared --enable-static --enable-libatomic --enable-threads=posix --enable-graphite --enable-fully-dynamic-string --enable-libstdcxx-filesystem-ts --enable-libstdcxx-time --disable-libstdcxx-pch --disable-libstdcxx-debug --enable-lto --enable-libgomp --disable-multilib --disable-rpath --disable-win32-registry --disable-nls --disable-werror --disable-symvers --with-libiconv --with-system-zlib --with-gmp=/mingw64 --with-mpfr=/mingw64 --with-mpc=/mingw64 --with-isl=/mingw64 --with-pkgversion='Rev10, Built by MSYS2 project' --with-bugurl=https://github.com/msys2/MINGW-packages/issues --with-gnu-as --with-gnu-ld --with-boot-ldflags='-pipe -Wl,--disable-dynamicbase -static-libstdc++ -static-libgcc' LDFLAGS_FOR_TARGET=-pipe --enable-linker-plugin-flags='LDFLAGS=-static-libstdc++\ -static-libgcc\ -pipe\ -Wl,--stack,12582912'`
	cc, err := parseConfig(input)
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := yaml.Marshal(cc)
		fmt.Printf("%s", string(buf))
	}

	// Output:
	// build: x86_64-w64-mingw32
	// disable-libstdcxx-debug: ""
	// disable-libstdcxx-pch: ""
	// disable-multilib: ""
	// disable-nls: ""
	// disable-rpath: ""
	// disable-symvers: ""
	// disable-werror: ""
	// disable-win32-registry: ""
	// enable-bootstrap: ""
	// enable-checking: release
	// enable-fully-dynamic-string: ""
	// enable-graphite: ""
	// enable-languages: c,lto,c++,fortran,ada,objc,obj-c++,jit
	// enable-libatomic: ""
	// enable-libgomp: ""
	// enable-libstdcxx-filesystem-ts: ""
	// enable-libstdcxx-time: ""
	// enable-linker-plugin-flags: LDFLAGS=-static-libstdc++\ -static-libgcc\ -pipe\ -Wl,--stack,12582912
	// enable-lto: ""
	// enable-shared: ""
	// enable-static: ""
	// enable-threads: posix
	// host: x86_64-w64-mingw32
	// libexecdir: /mingw64/lib
	// prefix: /mingw64
	// target: x86_64-w64-mingw32
	// with-arch: x86-64
	// with-boot-ldflags: -pipe -Wl,--disable-dynamicbase -static-libstdc++ -static-libgcc
	// with-bugurl: https://github.com/msys2/MINGW-packages/issues
	// with-gmp: /mingw64
	// with-gnu-as: ""
	// with-gnu-ld: ""
	// with-isl: /mingw64
	// with-libiconv: ""
	// with-local-prefix: /mingw64/local
	// with-mpc: /mingw64
	// with-mpfr: /mingw64
	// with-native-system-header-dir: /mingw64/include
	// with-pkgversion: Rev10, Built by MSYS2 project
	// with-system-zlib: ""
	// with-tune: generic
}
