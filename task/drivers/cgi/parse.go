package cgi

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/ghodss/yaml"
)

type TaskConfig struct {
	Spec Spec `json:"task"`
}

type Spec struct {
	Deps Deps `json:"deps"`
	Run  Run  `json:"run"`
}

// Deps represents the dependencies for different package managers
type Deps struct {
	// keeping these as structs instead of string lists so they can be extended
	// in the future if needed.
	Apt  []AptDep  `json:"apt"`
	Brew []BrewDep `json:"brew"`
}

// AptDep is a dependency for apt
type AptDep struct {
	Name string `json:"name"`
}

// BrewDep is a dependency for brew
type BrewDep struct {
	Name string `json:"name"`
}

// Run represents the run configuration
type Run struct {
	Go   *Go   `json:"go"`
	Bash *Bash `json:"bash"`
}

type Bash struct {
	Script string `json:"script"`
}

// Go represents the Go module configuration
type Go struct {
	Module string `json:"module"`
}

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*TaskConfig, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(TaskConfig)
	err = json.Unmarshal(b, out)
	return out, err
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*TaskConfig, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*TaskConfig, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*TaskConfig, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
