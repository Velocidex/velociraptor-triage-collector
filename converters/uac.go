// Convert UAC artifact format to standard form.
package converters

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Velocidex/velociraptor-triage-collector/api"
	"github.com/Velocidex/yaml/v2"
)

type UACArtifact struct {
	Description string   `json:"description"`
	SupportedOS []string `json:"supported_os"`
	Collector   string   `json:"collector"`
	Path        string   `json:"path"`
	PathPattern []string `json:"path_pattern"`
	NamePattern []string `json:"name_pattern"`
	FileType    []string `json:"file_type"`
	MaxFileSize uint64   `json:"max_file_size"`

	ExcludePathPattern  []string `json:"exclude_path_pattern"`
	ExcludeNamePattern  []string `json:"exclude_name_pattern"`
	ExcludeNologinUsers bool     `json:"exclude_nologin_users"`
	IgnoreDateRange     bool     `json:"ignore_date_range"`
	Permissions         []int64  `json:"permissions"`

	// Ignore
	MaxDepth int64 `json:"max_depth"`

	// Used by stat collector
	ExcludeFileSystem []string `json:"exclude_file_system"`
	OutputFile        string   `json:"output_file"`
	OutputDirectory   string   `json:"output_directory"`

	// Used by command collector
	Command                string `json:"command"`
	Condition              string `json:"condition"`
	Foreach                string `json:"foreach"`
	RedirectStderrToStdout bool   `json:"redirect_stderr_to_stdout"`
	IsFileList             bool   `json:"is_file_list"`
	CompressOutputFile     bool   `json:"compress_output_file"`

	// Used by the find collector
	NoGroup bool `json:"no_group"`
	NoUser  bool `json:"no_user"`
}

type UACRuleFile struct {
	Version     string        `json:"version"`
	Artifacts   []UACArtifact `json:"artifacts"`
	Description string        `json:"description"`

	// Ignored - used by stats collector.
	OutputDirectory string `json:"output_directory"`
	OutputFile      string `json:"output_file"`
	Condition       string `json:"condition"`
	Modifier        bool   `json:"modifier"`
}

func UACConvertFile(
	config *api.Config, filename string) (string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", err
	}

	res, err := UACConvert(config, filename, data)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getPrefix(a, b string) string {
	for idx := range b {
		if len(a) <= idx {
			return a
		}

		if a[idx] != b[idx] {
			return a[:idx]
		}
	}
	return a
}

func getDescription(rules *UACRuleFile) string {
	var description string
	for _, desc := range rules.Artifacts {
		if desc.Description != "" {
			commonPrefix := getPrefix(description, desc.Description)
			if commonPrefix == "" {
				commonPrefix = desc.Description
			}
			if len(commonPrefix) > 0 {
				description = commonPrefix
			}
		}
	}

	return strings.TrimSpace(description)
}

func UACConvert(
	config *api.Config, filename string, in []byte) ([]byte, error) {
	rules := &UACRuleFile{}
	err := yaml.UnmarshalStrict(in, rules)
	if err != nil {
		return nil, err
	}

	result := &api.TargetFile{
		Name:        makeTarget(filename),
		Description: getDescription(rules),
	}

	for _, artifact := range rules.Artifacts {
		if artifact.Collector != "file" {
			continue
		}

		result.Rules = append(result.Rules, &api.TargetRule{
			Name: sanitize(strings.TrimPrefix(artifact.Description, result.Description)),
			Glob: makeGlob(config, artifact),
		})
	}

	return yaml.Marshal(result)
}

func makeTarget(filename string) string {
	basename := filepath.Base(filename)
	basename = strings.Split(basename, ".")[0]
	return strings.Title(basename)
}

// Quote regex
var (
	quoteRegex     = regexp.MustCompile(`"([^/"]+?)"`)
	expansionRegex = regexp.MustCompile(`%([^/%]+?)%`)
	slashRegex     = regexp.MustCompile("/+")
)

func makeGlob(
	config *api.Config,
	artifact UACArtifact) string {
	base_path := artifact.Path

	// Remove useless quotes
	base_path = quoteRegex.ReplaceAllString(base_path, "$1")

	// Replace expansions
	base_path = expansionRegex.ReplaceAllStringFunc(
		base_path, func(in string) string {
			glob, pres := config.RegExToGlob[in]
			if pres {
				return glob
			}
			fmt.Printf("No glob substitution for expansion %v\n", in)
			return in
		})

	if len(artifact.PathPattern) > 1 {
		base_path += "/{" + strings.Join(artifact.PathPattern, ",") + "}"
	} else if len(artifact.PathPattern) == 1 {
		base_path += "/" + artifact.PathPattern[0]
	} else if len(artifact.NamePattern) > 1 {
		base_path += "/{" + strings.Join(artifact.NamePattern, ",") + "}"
	} else if len(artifact.NamePattern) == 1 {
		base_path += "/" + artifact.NamePattern[0]
	} else if len(artifact.FileType) > 0 {
		base_path += "/*"
	}

	base_path = slashRegex.ReplaceAllString(base_path, "/")

	return base_path
}

var (
	sanitizeRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func sanitize(in string) string {
	return strings.TrimSuffix(
		strings.TrimPrefix(sanitizeRegex.ReplaceAllString(in, "_"), "_"), "_")
}
