// Package nfpm provides ways to package programs in some linux packaging
// formats.
package nfpm

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	packagers = map[string]Packager{}
	lock      sync.Mutex
)

// Register a new packager for the given format
func Register(format string, p Packager) {
	lock.Lock()
	packagers[format] = p
	lock.Unlock()
}

// Get a packager for the given format
func Get(format string) (Packager, error) {
	p, ok := packagers[format]
	if !ok {
		return nil, fmt.Errorf("no packager registered for the format %s", format)
	}
	return p, nil
}

// Packager represents any packager implementation
type Packager interface {
	Package(info Info, w io.Writer) error
}

// Info contains information about the package
type Info struct {
	Name        string            `yaml:"name,omitempty"`
	Arch        string            `yaml:"arch,omitempty"`
	Platform    string            `yaml:"platform,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Section     string            `yaml:"section,omitempty"`
	Priority    string            `yaml:"priority,omitempty"`
	Replaces    []string          `yaml:"replaces,omitempty"`
	Provides    []string          `yaml:"provides,omitempty"`
	Depends     []string          `yaml:"depends,omitempty"`
	Recommends  []string          `yaml:"recommends,omitempty"`
	Suggests    []string          `yaml:"suggests,omitempty"`
	Conflicts   []string          `yaml:"conflicts,omitempty"`
	Maintainer  string            `yaml:"maintainer,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Vendor      string            `yaml:"vendor,omitempty"`
	Homepage    string            `yaml:"homepage,omitempty"`
	License     string            `yaml:"license,omitempty"`
	Bindir      string            `yaml:"bindir,omitempty"`
	Files       map[string]string `yaml:"files,omitempty"`
	ConfigFiles map[string]string `yaml:"config_files,omitempty"`
}

// WithDefaults set some sane defaults into the given Info
func WithDefaults(info Info) Info {
	if info.Bindir == "" {
		info.Bindir = "/usr/local/bin"
	}
	if info.Platform == "" {
		info.Platform = "linux"
	}
	info.Version = strings.TrimPrefix(info.Version, "v")
	return info
}
