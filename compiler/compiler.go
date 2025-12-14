package compiler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Velocidex/ordereddict"
	"github.com/Velocidex/velociraptor-triage-collector/api"
	"github.com/Velocidex/velociraptor-triage-collector/converters"
	"github.com/Velocidex/yaml/v2"
)

type Compiler struct {
	config_obj *api.Config
	targets    *ordereddict.Dict // map[string]*TargetFile

	template string

	logger *log.Logger

	// Dependent artifacts
	deps map[string]bool
}

// Load the targets from the directory recursively.
func (self *Compiler) LoadDirectory(
	compile_dir string,
	filter *regexp.Regexp,
	transformer api.Transformer) error {
	self.logger.Printf("Loading targets from directory %v", compile_dir)

	skip_lookup := make(map[string]bool)
	for _, name := range self.config_obj.SkipFiles {
		skip_lookup[name] = true
	}

	err := filepath.WalkDir(compile_dir,
		func(path string, d fs.DirEntry, err error) error {
			if !filter.MatchString(path) {
				return nil
			}

			if skip_lookup[d.Name()] {
				return nil
			}

			fd, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fd.Close()

			data, err := ioutil.ReadAll(fd)
			if err != nil {
				return err
			}

			transformed, err := transformer(self.config_obj, path, data)
			if err != nil {
				self.logger.Printf("Failed to load %v: %v", path, err)
				return err
			}

			// Allow each file to contain multiple rules.
			err = self.LoadRule(transformed, path)
			if err != nil {
				fmt.Printf("Rule %v: %v: %v\n", path,
					err, string(data))
				return err
			}

			return nil
		})

	return err
}

func (self *Compiler) LoadRule(data []byte, path string) error {
	self.logger.Printf("Loading target %v", path)

	target_file := &api.TargetFile{}
	err := yaml.UnmarshalStrict(data, target_file)
	if err != nil {
		return err
	}

	if target_file.Name == "" {
		target_file.Name = strings.Split(filepath.Base(path), ".")[0]
	}

	if len(target_file.Rules) > 0 || len(target_file.Targets) > 0 {
		self.targets.Set(target_file.Name, target_file)
	}
	return nil
}

func (self *Compiler) clearLagcyTargetFile(tf *api.TargetFile) {
	// Support legacy KapeFiles descriptors. We rename the field
	// to Rules.
	tf.Rules = append(tf.Rules, tf.Targets...)
	tf.Targets = nil
	tf.Id = ""
	tf.RecreateDirectories = false
	tf.Version = ""
}

func (self *Compiler) validate() error {
	if self.config_obj.PathSep == "" {
		self.config_obj.PathSep = "/"
	}

	for _, t_file := range self.targets.Values() {
		tf := t_file.(*api.TargetFile)
		self.clearLagcyTargetFile(tf)

		for _, t := range tf.Rules {
			err := self.ValidateRule(t, tf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Compiler) SaveState(path string) error {
	serialized, err := json.MarshalIndent(self.targets, " ", " ")
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Write(serialized)
	return err
}

func LoadConfig(path string) (*api.Config, error) {
	config_obj := &api.Config{}

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(data, &config_obj)
	if err != nil {
		return nil, err
	}

	return config_obj, nil
}

func (self *Compiler) loadConfig(path string) error {
	self.logger.Printf("Loading config from %v", path)

	config_obj, err := LoadConfig(path)
	if err != nil {
		return err
	}
	self.config_obj = config_obj

	target_regex := self.config_obj.TargetRegex
	if target_regex == "" {
		target_regex = "(.tkape|.yaml)$"
	}

	filter, err := regexp.Compile(target_regex)
	if err != nil {
		return err
	}

	var transformer api.Transformer

	switch self.config_obj.Transformer {
	case "":
		transformer = func(
			config *api.Config, filename string, in []byte) ([]byte, error) {
			return in, nil
		}
	case "uac":
		transformer = converters.UACConvert
	}

	// Load the targets
	for _, target_dir := range self.config_obj.TargetDirectories {
		err = self.LoadDirectory(target_dir, filter, transformer)
		if err != nil {
			return err
		}
	}

	if self.config_obj.MakeAllTarget {
		new_target := &api.TargetFile{Name: "_All"}
		for _, item := range self.targets.Items() {
			tf, ok := item.Value.(*api.TargetFile)
			if !ok {
				continue
			}
			new_target.Targets = append(new_target.Targets, &api.TargetRule{
				Name: tf.Name,
				Ref:  tf.Name,
			})
		}
		self.targets.Set("_All", new_target)
	}

	// Load the template
	fd, err := os.Open(self.config_obj.ArtifactTemplate)
	if err != nil {
		return err
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	self.template = string(data)

	return self.validate()
}

func (self *Compiler) Run() error {
	for _, output := range self.config_obj.Output {
		artifact, err := self.GetArtifact()
		if err != nil {
			return err
		}

		fd, err := os.OpenFile(output,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer fd.Close()

		if strings.HasSuffix(output, ".zip") {
			self.logger.Printf("Generating Zip artifact pack into %v", output)
			zip := zip.NewWriter(fd)
			defer zip.Close()

			out_fd, err := zip.Create(self.config_obj.Name + ".yaml")
			if err != nil {
				return err
			}
			out_fd.Write([]byte(artifact))
		} else {

			self.logger.Printf("Generating YAML artifact pack into %v", output)
			fd.Write([]byte(artifact))
		}
	}

	if self.config_obj.StateFile != "" {
		err := self.SaveState(self.config_obj.StateFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCompiler(config_path string, logger *log.Logger) (*Compiler, error) {
	res := &Compiler{
		targets: ordereddict.NewDict(), // make(map[string]*TargetFile),
		logger:  logger,
		deps:    make(map[string]bool),
	}

	return res, res.loadConfig(config_path)
}
