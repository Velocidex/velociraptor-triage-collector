package compiler

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/Velocidex/velociraptor-triage-collector/api"
)

type RuleCSV struct {
	// The Target name this came from.
	Target      string
	Name        string
	Description string
	Glob        string
	Ref         string

	VQL string
}

type TargetCSV struct {
	Name        string
	Description string
	Preamble    string
}

type ArtifactContent struct {
	Time        string
	Commit      string
	Rules       []*RuleCSV
	TargetFiles []*TargetCSV

	Config *api.Config

	Dependencies map[string]bool
}

func readFile(args ...interface{}) interface{} {
	result := ""

	for _, arg := range args {
		path, ok := arg.(string)
		if !ok {
			continue
		}

		fd, err := os.Open(path)
		if err != nil {
			continue
		}
		defer fd.Close()

		data, err := ioutil.ReadAll(fd)
		if err != nil {
			continue
		}

		result += string(data)
	}

	return result
}

func indentTemplate(args ...interface{}) interface{} {
	if len(args) != 2 {
		return ""
	}

	template, ok := args[0].(string)
	if !ok {
		return ""
	}

	indent_size, ok := args[1].(int)
	if !ok {
		return template
	}

	return indent(template, indent_size)
}

func expandTemplate(template_path string, params *ArtifactContent) string {
	fd, err := os.Open(template_path)
	if err != nil {
		return fmt.Sprintf("Unable to open %v: %v", template_path, err)
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return fmt.Sprintf("Unable to open %v: %v", template_path, err)
	}

	res, err := calculateTemplate(string(data), params)

	if err != nil {
		return fmt.Sprintf("Expanding template %v: %v", template_path, err)
	}
	return res
}

func calculateTemplate(template_str string, params *ArtifactContent) (string, error) {
	var templ *template.Template
	var err error

	funcMap := template.FuncMap{
		"Indent":   indentTemplate,
		"ReadFile": readFile,

		"Template": func(template string) string {
			return expandTemplate(template, params)
		},

		// Compress a template into base64
		"Compress": func(name string, args interface{}) string {
			b := &bytes.Buffer{}
			err := templ.ExecuteTemplate(b, name, args)
			if err != nil {
				return fmt.Sprintf("<%v>", err)
			}

			// Compress the string and encode as base64
			bc := &bytes.Buffer{}
			gz, err := gzip.NewWriterLevel(bc, 9)
			if err != nil {
				return fmt.Sprintf("<%v>", err)
			}

			gz.Write(b.Bytes())
			gz.Close()

			enc := &bytes.Buffer{}
			encoder := base64.NewEncoder(base64.StdEncoding, enc)
			encoder.Write(bc.Bytes())
			encoder.Close()
			return string(enc.Bytes())
		},
	}

	templ, err = template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(funcMap).Parse(template_str)
	if err != nil {
		return "", err
	}

	b := &bytes.Buffer{}
	err = templ.Execute(b, params)
	if err != nil {
		return "", err
	}

	return string(b.Bytes()), nil
}

func (self *Compiler) GetCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (self *Compiler) GetArtifact() (string, error) {
	params := &ArtifactContent{
		Time:         time.Now().UTC().Format(time.RFC3339),
		Commit:       self.GetCommit(),
		Config:       self.config_obj,
		Dependencies: self.deps,
	}

	for _, target_file_any := range self.targets.Values() {
		target_file := target_file_any.(*api.TargetFile)
		params.TargetFiles = append(params.TargetFiles,
			&TargetCSV{
				Name:        sanitize(target_file.Name),
				Description: target_file.Description,
				Preamble:    target_file.Preamble,
			})

		for _, t := range target_file.Rules {
			params.Rules = append(params.Rules, &RuleCSV{
				Target:      sanitize(target_file.Name),
				Description: t.Comment,
				Name:        sanitize(t.Name),
				Glob:        t.Glob,
				Ref:         sanitize(t.Ref),
				VQL:         t.VQL,
			})

			if t.VQL != "" {
				self.GetDependentArtifacts(t.VQL)
			}
		}
	}

	sort.Slice(params.Rules, func(i, j int) bool {
		key1 := params.Rules[i].Target + params.Rules[i].Name
		key2 := params.Rules[j].Target + params.Rules[j].Name
		return key1 < key2
	})

	sort.Slice(params.TargetFiles, func(i, j int) bool {
		key1 := strings.ReplaceAll(params.TargetFiles[i].Name, "_", " ")
		key2 := strings.ReplaceAll(params.TargetFiles[j].Name, "_", " ")
		return key1 < key2
	})

	return calculateTemplate(self.template, params)
}

var (
	artifact_in_query_regex = regexp.MustCompile(`Artifact\.([^\s\(]+)\(`)
)

func (self *Compiler) GetDependentArtifacts(query string) {
	for _, hit := range artifact_in_query_regex.FindAllStringSubmatch(query, -1) {
		artifact_name := hit[1]
		self.deps[artifact_name] = true
	}
}

func indent(in string, indent int) string {
	indent_str := strings.Repeat(" ", indent)
	lines := strings.Split(in, "\n")
	result := []string{}
	for _, l := range lines {
		result = append(result, indent_str+l)
	}
	return strings.Join(result, "\n")
}

var (
	sanitizeRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func sanitize(in string) string {
	if len(in) > 0 && in[0] == '$' {
		in = in[1:]
	}
	return sanitizeRegex.ReplaceAllString(in, "_")
}
