package v1

import (
	"path"
	"strings"

	utilsstrings "github.com/yamajik/kess/utils/strings"
)

// NamedVersion bulabula
type NamedVersion struct {
	Name    string
	Version string
}

// Constants bulabula
const (
	VarSeparator     = "_"
	NameSeparator    = "-"
	VersionSeparator = "."
	LatestVersion    = "latest"
)

// NamedVersionFromString bulabula
func NamedVersionFromString(s string) NamedVersion {
	var (
		slice        = strings.Split(s, "-")
		namedversion = NamedVersion{
			Name:    slice[0],
			Version: LatestVersion,
		}
	)
	if len(slice) > 1 {
		namedversion.Version = slice[1]
	}
	return namedversion
}

// String bulabula
func (v NamedVersion) String() string {
	return strings.Join([]string{v.Name, v.Version}, NameSeparator)
}

// Path bulabula
func (v NamedVersion) Path() string {
	return path.Join(v.Name, v.Version)
}

// NameVar bulabula
func (v NamedVersion) NameVar() string {
	return strings.Replace(v.Name, NameSeparator, VarSeparator, -1)
}

// VersionVar bulabula
func (v NamedVersion) VersionVar() string {
	return strings.Replace(v.Version, VersionSeparator, VarSeparator, -1)
}

// FileName bulabula
func (v NamedVersion) FileName() string {
	return strings.Join([]string{v.NameVar(), v.VersionVar()}, VarSeparator)
}

// Map bulabula
func (v NamedVersion) Map() map[string]interface{} {
	return map[string]interface{}{
		"Name":    v.Name,
		"Version": v.Version,
	}
}

// Format bulabula
func (v NamedVersion) Format(s string) string {
	return utilsstrings.Format(s, v.Map())
}
