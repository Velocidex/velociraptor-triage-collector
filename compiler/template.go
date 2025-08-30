package compiler

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Velocidex/velociraptor-triage-collector/api"
)

type RuleCSV struct {
	// The Target name this came from.
	Target      string
	Name        string
	Description string
	Glob        string
	Ref         string
}

type TargetCSV struct {
	Name        string
	Description string
}

type ArtifactContent struct {
	Time        string
	Commit      string
	Rules       []*RuleCSV
	TargetFiles []*TargetCSV
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

func calculateTemplate(template_str string, params *ArtifactContent) (string, error) {
	templ, err := template.New("").Funcs(
		template.FuncMap{
			"Indent":   indentTemplate,
			"ReadFile": readFile,
		}).Parse(template_str)
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
		Time:   time.Now().UTC().Format(time.RFC3339),
		Commit: self.GetCommit(),
	}

	for _, target_file_any := range self.targets.Values() {
		target_file := target_file_any.(*api.TargetFile)
		params.TargetFiles = append(params.TargetFiles,
			&TargetCSV{
				Name:        sanitize(target_file.Name),
				Description: target_file.Description,
			})

		for _, t := range target_file.Rules {
			params.Rules = append(params.Rules, &RuleCSV{
				Target: sanitize(target_file.Name),
				Name:   sanitize(t.Name),
				Glob:   t.Glob,
				Ref:    sanitize(t.Ref),
			})
		}
	}

	sort.Slice(params.Rules, func(i, j int) bool {
		key1 := params.Rules[i].Target + params.Rules[i].Name
		key2 := params.Rules[j].Target + params.Rules[j].Name
		return key1 < key2
	})

	return calculateTemplate(self.template, params)
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
	return sanitizeRegex.ReplaceAllString(in, "_")
}
