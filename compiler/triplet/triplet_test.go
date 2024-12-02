package triplet

import (
	"reflect"
	"testing"
)

func TestParseFull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Full
	}{
		{
			name:  "x86_64 linux gnu",
			input: "x86_64-linux-gnu",
			expected: Full{
				Target: Target{
					Arch: "x64",
					OS:   "linux",
					ABI:  "elf",
					LibC: "glibc",
				},
				Original: "x86_64-linux-gnu",
				Vendors:  []string{"gnu"},
			},
		},
		{
			name:  "aarch64 linux gnu",
			input: "aarch64-linux-gnu",
			expected: Full{
				Target: Target{
					Arch: "arm64",
					OS:   "linux",
					ABI:  "elf",
					LibC: "glibc",
				},
				Original: "aarch64-linux-gnu",
				Vendors:  []string{"gnu"},
			},
		},
		{
			name:  "x86_64 windows msvc",
			input: "x86_64-windows-msvc",
			expected: Full{
				Target: Target{
					Arch: "x64",
					OS:   "windows",
					ABI:  "pe",
					LibC: "msvcrt",
				},
				Original: "x86_64-windows-msvc",
				Vendors:  []string{},
			},
		},
		{
			name:  "arm darwin",
			input: "arm-darwin",
			expected: Full{
				Target: Target{
					Arch: "arm",
					OS:   "darwin",
					ABI:  "marcho",
					LibC: "unknown",
				},
				Original: "arm-darwin",
				Vendors:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFull(tt.input)
			if err != nil {
				t.Fatalf("ParseFull(%q) unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseFull(%q)\ngot  = %#v\nwant = %#v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTarget_Match(t *testing.T) {
	tests := []struct {
		name     string
		target   Target
		other    Target
		expected bool
	}{
		{
			name: "exact match",
			target: Target{
				Arch: "x64",
				OS:   "linux",
				ABI:  "elf",
				LibC: "glibc",
			},
			other: Target{
				Arch: "x64",
				OS:   "linux",
				ABI:  "elf",
				LibC: "glibc",
			},
			expected: true,
		},
		{
			name: "partial match - empty fields",
			target: Target{
				Arch: "x64",
				OS:   "linux",
			},
			other: Target{
				Arch: "x64",
				OS:   "linux",
				ABI:  "elf",
				LibC: "glibc",
			},
			expected: true,
		},
		{
			name: "no match - different arch",
			target: Target{
				Arch: "x64",
				OS:   "linux",
			},
			other: Target{
				Arch: "arm64",
				OS:   "linux",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.Match(tt.other)
			if got != tt.expected {
				t.Errorf("Target.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseArch(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldMatch bool
	}{
		{"x86_64", "x64", true},
		{"amd64", "x64", true},
		{"i386", "x32", true},
		{"i686", "x32", true},
		{"aarch64", "arm64", true},
		{"arm", "arm", true},
		{"riscv64", "riscv64", true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseArch(tt.input)
			if ok != tt.shouldMatch {
				t.Errorf("ParseArch(%q) match = %v, want %v", tt.input, ok, tt.shouldMatch)
			}
			if tt.shouldMatch && got != tt.expected {
				t.Errorf("ParseArch(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseOS(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldMatch bool
	}{
		{"linux", "linux", true},
		{"linux-gnu", "linux", true},
		{"windows", "windows", true},
		{"mingw32", "windows", true},
		{"darwin", "darwin", true},
		{"darwin20", "darwin", true},
		{"freebsd", "freebsd", true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseOS(tt.input)
			if ok != tt.shouldMatch {
				t.Errorf("ParseOS(%q) match = %v, want %v", tt.input, ok, tt.shouldMatch)
			}
			if tt.shouldMatch && got != tt.expected {
				t.Errorf("ParseOS(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseABI(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldMatch bool
	}{
		{"elf", "elf", true},
		{"eabi", "eabi", true},
		{"mingw32", "pe", true},
		{"msvc", "pe", true},
		{"linux", "elf", true},
		{"darwin", "marcho", true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseABI(tt.input)
			if ok != tt.shouldMatch {
				t.Errorf("ParseABI(%q) match = %v, want %v", tt.input, ok, tt.shouldMatch)
			}
			if tt.shouldMatch && got != tt.expected {
				t.Errorf("ParseABI(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseLibC(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldMatch bool
	}{
		{"gnu", "glibc", true},
		{"glibc", "glibc", true},
		{"musl", "musl", true},
		{"mingw", "mingw", true},
		{"msvc", "msvcrt", true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseLibC(tt.input)
			if ok != tt.shouldMatch {
				t.Errorf("ParseLibC(%q) match = %v, want %v", tt.input, ok, tt.shouldMatch)
			}
			if tt.shouldMatch && got != tt.expected {
				t.Errorf("ParseLibC(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTarget_String(t *testing.T) {
	tests := []struct {
		name     string
		target   Target
		expected string
	}{
		{
			name: "full target",
			target: Target{
				Arch: "x64",
				OS:   "linux",
				ABI:  "elf",
				LibC: "glibc",
			},
			expected: "x64-linux-elf-glibc",
		},
		{
			name: "partial target",
			target: Target{
				Arch: "arm64",
				OS:   "darwin",
				ABI:  "marcho",
			},
			expected: "arm64-darwin-marcho",
		},
		{
			name: "minimal target",
			target: Target{
				Arch: "x64",
				OS:   "windows",
			},
			expected: "x64-windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.String()
			if got != tt.expected {
				t.Errorf("Target.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
