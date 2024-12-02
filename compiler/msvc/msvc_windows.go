package msvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-build/compiler/triplet"
	"github.com/adnsv/go-utils/filesystem"
)

const vswhereSubpath = "Microsoft Visual Studio/Installer/vswhere.exe"

// DiscoverInstallations finds all MSVC installations using multiple methods
func DiscoverInstallations(feedback func(string)) ([]*Installation, error) {
	if feedback != nil {
		feedback("discovering msvc installations")
	}

	installations := []*Installation{}

	// Try vswhere first
	if vsInstalls, err := discoverViaVSWhere(feedback); err == nil {
		installations = append(installations, vsInstalls...)
	}

	// Try environment variables
	if envInstalls, err := discoverViaEnvironment(feedback); err == nil {
		installations = append(installations, envInstalls...)
	}

	// Try standalone installations
	if standaloneInstalls, err := discoverStandalone(feedback); err == nil {
		installations = append(installations, standaloneInstalls...)
	}

	// Deduplicate and sort installations
	installations = deduplicateInstallations(installations)

	if feedback != nil {
		feedback(fmt.Sprintf("found %d msvc installation(s)", len(installations)))
	}

	return installations, nil
}

// discoverViaVSWhere discovers Visual Studio installations using vswhere utility
func discoverViaVSWhere(feedback func(string)) ([]*Installation, error) {
	vswherePath := findVSWhere()
	if vswherePath == "" {
		return nil, errors.New("failed to find vswhere.exe")
	}

	if feedback != nil {
		feedback(fmt.Sprintf("using vswhere utility: %s", vswherePath))
	}

	cmd := exec.Command(vswherePath, "-all", "-format", "json", "-products", "*", "-legacy", "-prerelease")
	buf, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	type vsWhereInstallation struct {
		InstanceID          string `json:"instanceId"`
		DisplayName         string `json:"displayName"`
		InstallationPath    string `json:"installationPath"`
		InstallationVersion string `json:"installationVersion"`
		Description         string `json:"description"`
		IsPrerelease        bool   `json:"isPrerelease"`
	}

	tmp := []*vsWhereInstallation{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return nil, fmt.Errorf("failed to parse vswhere output: %s", err)
	}

	installations := make([]*Installation, 0, len(tmp))
	for _, i := range tmp {
		inst := &Installation{
			InstanceID:          i.InstanceID,
			DisplayName:         i.DisplayName,
			InstallationPath:    filepath.ToSlash(i.InstallationPath),
			InstallationVersion: i.InstallationVersion,
			Description:         i.Description,
			IsPrerelease:        i.IsPrerelease,
			DiscoveryMethod:     "vswhere",
		}

		// Detect all toolset versions
		inst.ToolsetVersions = detectToolsetVersions(inst.InstallationPath, feedback)
		installations = append(installations, inst)
	}

	return installations, nil
}

// discoverViaEnvironment discovers Visual Studio installations using environment variables
func discoverViaEnvironment(feedback func(string)) ([]*Installation, error) {
	if feedback != nil {
		feedback("discovering msvc installations via environment variables")
	}

	installations := []*Installation{}

	// Check CL environment variable first
	if cl := os.Getenv("CL"); cl != "" {
		if inst := validateCLCompiler(cl, feedback); inst != nil {
			installations = append(installations, inst)
		}
	}

	// Check VS* environment variables
	for _, env := range []string{"VS140COMNTOOLS", "VS120COMNTOOLS", "VS110COMNTOOLS"} {
		if path := os.Getenv(env); path != "" {
			if inst := validateVSEnvironment(env, path, feedback); inst != nil {
				installations = append(installations, inst)
			}
		}
	}

	return installations, nil
}

// discoverStandalone discovers standalone MSVC installations
func discoverStandalone(feedback func(string)) ([]*Installation, error) {
	if feedback != nil {
		feedback("discovering standalone msvc installations")
	}

	installations := []*Installation{}
	commonPaths := []string{
		`C:\Program Files (x86)\Microsoft Visual Studio`,
		`C:\Program Files\Microsoft Visual Studio`,
	}

	for _, basePath := range commonPaths {
		if !filesystem.DirExists(basePath) {
			continue
		}

		// Look for version-specific directories
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			path := filepath.Join(basePath, entry.Name())
			if inst := validateStandaloneMSVC(path, feedback); inst != nil {
				installations = append(installations, inst)
			}
		}
	}

	return installations, nil
}

// findVSWhere looks for vswhere.exe in standard locations
func findVSWhere() string {
	paths := []string{}
	addPath := func(s string) {
		for _, p := range paths {
			if p == s {
				return
			}
		}
		paths = append(paths, s)
	}

	if pf := os.Getenv("ProgramFiles(x86)"); filesystem.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if pf := os.Getenv("ProgramFiles"); filesystem.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if filesystem.DirExists("C:\\Program Files (x86)") {
		addPath(filepath.Join("C:\\Program Files (x86)", vswhereSubpath))
	}
	addPath("vswhere.exe")

	for _, path := range paths {
		if filesystem.FileExists(path) {
			return path
		}
	}
	return ""
}

// detectToolsetVersions finds all available toolset versions in a VS installation
func detectToolsetVersions(installPath string, feedback func(string)) []ToolsetVersion {
	versions := []ToolsetVersion{}

	// Check default toolset version
	defaultVersion := filepath.Join(installPath, "VC", "Auxiliary", "Build", "Microsoft.VCToolsVersion.default.txt")
	var defaultVer string
	if data, err := os.ReadFile(defaultVersion); err == nil {
		defaultVer = strings.TrimSpace(string(data))
		versions = append(versions, ToolsetVersion{
			Version:   defaultVer,
			Path:      filepath.Join(installPath, "VC", "Tools", "MSVC", defaultVer),
			IsDefault: true,
		})
	}

	// Check for additional toolsets
	toolsetsPath := filepath.Join(installPath, "VC", "Tools", "MSVC")
	if entries, err := os.ReadDir(toolsetsPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == defaultVer {
				continue
			}
			versions = append(versions, ToolsetVersion{
				Version: entry.Name(),
				Path:    filepath.Join(toolsetsPath, entry.Name()),
			})
		}
	}

	// Sort versions, latest first
	sort.Slice(versions, func(i, j int) bool {
		v1, e1 := ParseVersionQuad(versions[i].Version)
		v2, e2 := ParseVersionQuad(versions[j].Version)
		if e1 == nil && e2 == nil {
			return v1.Compare(v2) > 0
		}
		return versions[i].Version > versions[j].Version
	})

	return versions
}

// validateCLCompiler validates a CL compiler path and creates an Installation if valid
func validateCLCompiler(clPath string, feedback func(string)) *Installation {
	if !filesystem.FileExists(clPath) {
		return nil
	}

	// Try to get version information
	ver, _, err := QueryVersion(clPath)
	if err != nil {
		return nil
	}

	inst := &Installation{
		DisplayName:         fmt.Sprintf("MSVC Compiler %s", ver),
		InstallationPath:    filepath.Dir(clPath),
		InstallationVersion: ver,
		DiscoveryMethod:     "env",
		ToolsetVersions: []ToolsetVersion{{
			Version:   ver,
			Path:      filepath.Dir(clPath),
			IsDefault: true,
		}},
	}

	return inst
}

// validateVSEnvironment validates a VS environment variable and creates an Installation if valid
func validateVSEnvironment(envName, path string, feedback func(string)) *Installation {
	// VS*COMNTOOLS points to Common7/Tools, we need to go up two levels
	vsPath := filepath.Dir(filepath.Dir(path))
	if !filesystem.DirExists(vsPath) {
		return nil
	}

	// Try to find version from the environment variable name
	ver := strings.TrimPrefix(envName, "VS")
	ver = strings.TrimSuffix(ver, "COMNTOOLS")
	if ver == "" {
		return nil
	}

	majorVer := ver[:len(ver)-2] // e.g., "140" -> "14"
	minorVer := ver[len(ver)-2:] // e.g., "140" -> "0"

	inst := &Installation{
		DisplayName:         fmt.Sprintf("Visual Studio %s.%s", majorVer, minorVer),
		InstallationPath:    vsPath,
		InstallationVersion: fmt.Sprintf("%s.%s", majorVer, minorVer),
		DiscoveryMethod:     "env",
	}

	// Detect toolset versions
	inst.ToolsetVersions = detectToolsetVersions(vsPath, feedback)
	return inst
}

// validateStandaloneMSVC validates a standalone MSVC installation path
func validateStandaloneMSVC(path string, feedback func(string)) *Installation {
	vcvarsPath := filepath.Join(path, "VC", "vcvarsall.bat")
	if !filesystem.FileExists(vcvarsPath) {
		return nil
	}

	// Try to determine version from directory structure
	dirName := filepath.Base(path)
	inst := &Installation{
		DisplayName:         fmt.Sprintf("Standalone MSVC (%s)", dirName),
		InstallationPath:    path,
		InstallationVersion: dirName,
		DiscoveryMethod:     "standalone",
	}

	// Detect toolset versions
	inst.ToolsetVersions = detectToolsetVersions(path, feedback)
	return inst
}

// deduplicateInstallations removes duplicate installations based on path
func deduplicateInstallations(installations []*Installation) []*Installation {
	seen := make(map[string]bool)
	result := []*Installation{}

	for _, inst := range installations {
		if !seen[inst.InstallationPath] {
			seen[inst.InstallationPath] = true
			result = append(result, inst)
		}
	}

	// Sort installations by version, latest first
	sort.SliceStable(result, func(i, j int) bool {
		v1, e1 := ParseVersionQuad(result[i].InstallationVersion)
		v2, e2 := ParseVersionQuad(result[j].InstallationVersion)
		if e1 == nil && e2 == nil {
			return v1.Compare(v2) > 0
		}
		return result[i].InstallationVersion > result[j].InstallationVersion
	})

	return result
}

// getSupportedArchitectures returns a list of supported architecture configurations
func getSupportedArchitectures() []ArchitectureSpec {
	specs := []ArchitectureSpec{}
	baseArchs := []string{"x86", "amd64", "arm", "arm64"}
	hostArch := runtime.GOARCH

	// Add native compilation
	for _, arch := range baseArchs {
		specs = append(specs, ArchitectureSpec{
			Name:       arch,
			HostArch:   hostArch,
			TargetArch: arch,
		})
	}

	// Add cross-compilation if supported
	// Note: MSVC cross-compilation is only supported on AMD64 hosts because:
	// 1. Only 64-bit Windows can host the complete set of MSVC cross-compilers
	// 2. MSVC toolchain naming (e.g. amd64_x86, amd64_arm) assumes AMD64 host
	// 3. Visual Studio's cross-compilation support is designed around AMD64 hosts
	if hostArch == "amd64" {
		for _, arch := range baseArchs {
			if arch != hostArch {
				specs = append(specs, ArchitectureSpec{
					Name:          fmt.Sprintf("%s_%s", hostArch, arch),
					HostArch:      hostArch,
					TargetArch:    arch,
					CrossCompiler: true,
				})
			}
		}
	}

	return specs
}

// TestArches tests which architectures are supported by the installation
func TestArches(inst *Installation, feedback func(string)) []*toolchain.Chain {
	toolchains := []*toolchain.Chain{}
	specs := getSupportedArchitectures()

	// Find the latest toolset version
	var latestToolset *ToolsetVersion
	if len(inst.ToolsetVersions) > 0 {
		latestToolset = &inst.ToolsetVersions[0]
	}

	if latestToolset == nil {
		if feedback != nil {
			feedback("ERROR: no toolset versions found")
		}
		return nil
	}

	commonDir := filepath.Join(inst.InstallationPath, "Common7", "tools")
	devbat := filepath.Join(inst.InstallationPath, "VC", "Auxiliary", "Build", "vcvarsall.bat")

	if !filesystem.FileExists(devbat) {
		devbat = filepath.Join(inst.InstallationPath, "VC", "vcvarsall.bat")
		if !filesystem.FileExists(devbat) {
			if feedback != nil {
				feedback("ERROR: failed to locate vcvarsall.bat file")
			}
			return nil
		}
	}

	if feedback != nil {
		feedback(fmt.Sprintf("using devbat: %s", devbat))
	}

	// Test each architecture configuration
	var wg sync.WaitGroup
	results := make([]map[string]string, len(specs))

	for i, spec := range specs {
		wg.Add(1)
		go func(i int, spec ArchitectureSpec) {
			defer wg.Done()
			v, err := CollectBatVars(devbat, spec.Name, latestToolset.Version, commonDir, feedback)
			if err == nil {
				results[i] = v
			}
		}(i, spec)
	}
	wg.Wait()

	// Process results
	for i, spec := range specs {
		vars := results[i]
		if len(vars) == 0 {
			continue
		}

		if feedback != nil {
			feedback(fmt.Sprintf("%s: architecture %s - supported", inst.DisplayName, spec.Name))
		}

		tt := triplet.Full{
			Target: triplet.Target{
				Arch: triplet.NormalizeArch(spec.TargetArch),
				OS:   "windows",
				ABI:  "pe",
				LibC: "msvcrt",
			}}

		tc := &toolchain.Chain{
			Compiler:            "msvc",
			Implementation:      "msvc",
			FullVersion:         fmt.Sprintf("%s - %s - %s", inst.DisplayName, tt.Arch, inst.InstallationVersion),
			Version:             inst.InstallationVersion,
			Target:              tt,
			InstalledDir:        filepath.ToSlash(inst.InstallationPath),
			VisualStudioID:      inst.InstanceID,
			VisualStudioArch:    spec.Name,
			VisualStudioVersion: vars["VISUALSTUDIOVERSION"],
			UCRTVersion:         vars["UCRTVERSION"],
			ToolsetVersion:      latestToolset.Version,
			Tools:               toolchain.Toolset{},
		}

		if s := vars["WINDOWSSDKVERSION"]; s != "" {
			tc.WindowsSDKVersion = strings.TrimRight(s, `\`)
		}
		if incs := filepath.SplitList(vars["INCLUDE"]); len(incs) > 0 {
			for _, v := range incs {
				if v != "" && filesystem.DirExists(v) {
					tc.CCIncludeDirs = append(tc.CCIncludeDirs, filepath.ToSlash(v))
				}
			}
		}
		tc.CXXIncludeDirs = tc.CCIncludeDirs

		if libs := strings.Split(vars["LIB"], ";"); len(libs) > 0 {
			for _, v := range libs {
				if v != "" {
					tc.LibraryDirs = append(tc.LibraryDirs, filepath.ToSlash(v))
				}
			}
		}

		searchPaths := filepath.SplitList(vars["PATH"])
		for _, path := range searchPaths {
			for name, tool := range ToolNames {
				fn := filepath.Join(path, name+".exe")
				if filesystem.FileExists(fn) {
					fn = filepath.ToSlash(fn)
					tc.Tools[tool] = toolchain.ToolPath(fn)
					if tool == toolchain.CXXCompiler {
						tc.Tools[toolchain.CCompiler] = toolchain.ToolPath(fn)
					}
				}
			}
		}

		for env, tool := range ToolEnvs {
			if fn, ok := vars[env]; ok {
				if filesystem.FileExists(fn) {
					fn = filepath.ToSlash(fn)
					tc.Tools[tool] = toolchain.ToolPath(fn)
					if tool == toolchain.CXXCompiler {
						tc.Tools[toolchain.CCompiler] = toolchain.ToolPath(fn)
					}
				}
			}
		}

		for k, v := range vars {
			tc.Environment = append(tc.Environment, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tc.Environment)

		toolchains = append(toolchains, tc)
	}

	return toolchains
}

func CollectBatVars(devbat string, arg string, majorVer string, commonDir string, feedback func(string)) (map[string]string, error) {
	ret := map[string]string{}
	fn := "test.bat"
	batfname := "vs-cmt-" + fn
	envfname := batfname + ".env"
	bat := []string{
		`@echo off`,
		`cd /d "%~dp0"`,
		`set "VS` + majorVer + `0COMNTOOLS=` + commonDir + `"`,
		`call "` + devbat + `" ` + arg + ` || exit`,
	}
	for _, v := range msvcEnvVarsExtended {
		bat = append(bat, fmt.Sprintf("echo %s := %%%s%% >> %s\n", v, v, envfname))
	}
	tmpdir, err := os.MkdirTemp("", "bushido")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)
	batpath := filepath.Join(tmpdir, batfname)
	envpath := filepath.Join(tmpdir, envfname)
	err = os.WriteFile(batpath, []byte(strings.Join(bat, "\r\n")), 0777)
	if err != nil {
		return nil, err
	}

	err = exec.Command("cmd", "/C", batpath).Run()
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(envpath)
	if err != nil {
		return nil, err
	}

	if feedback != nil {
		feedback(fmt.Sprintf("- calling devbat %s", arg))
	}
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		if feedback != nil {
			feedback(fmt.Sprintf("> %s", line))
		}
		p := strings.Index(line, ":=")
		if p > 0 {
			n := strings.TrimSpace(line[:p])
			v := strings.TrimSpace(line[p+2:])
			ret[n] = v
			//if feedback != nil {
			//	feedback(fmt.Sprintf("    - %s=%s", n, v))
			//}
		}
	}
	if ret["INCLUDE"] == "" {
		return nil, fmt.Errorf("invalid batch output, can't find INCLUDE entry")
	}

	return ret, nil
}

func DiscoverToolchains(feedback func(string)) []*toolchain.Chain {
	ret := []*toolchain.Chain{}
	msvcs, err := DiscoverInstallations(feedback)
	if err != nil {
		if feedback != nil {
			feedback(err.Error())
		}
		return nil
	}
	for _, msvc := range msvcs {
		ret = append(ret, TestArches(msvc, feedback)...)
	}
	return ret
}
