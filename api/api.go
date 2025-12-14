package api

// Describe a collection target.
//
// This is based on
// https://ericzimmerman.github.io/KapeDocs/#!Pages\2.1-Targets.md#targets
// but extended and clarified.
type TargetRule struct {
	Name     string `json:"Name,omitempty"`
	Category string `json:"Category,omitempty"`
	Comment  string `json:"Comment,omitempty"`

	// A rule may reference another TargetFile which will have all its
	// targets expanded at runtime.
	Ref string `json:"Ref,omitempty"`

	// A direct glob to the files that will be included. This is the
	// preferred field for newer targets instead of the fields below.
	Glob string `json:"Glob,omitempty"`

	// Rules may contain a full VQL query which will generate a list
	// of files to upload. The query must generate the following
	// columns:
	// - OSPath: This will be the file that is uploaded.
	// - Size: The size of the file to be uploaded.
	// - Btime, Ctime, Mtime, Atime: The times of the uploaded file.
	// - Accessor: The accessor to use to upload the file.
	VQL string `json:"VQL,omitempty"`

	// The KapeFiles was the initial inspiration for this collector,
	// but the meaning of these fields is not well documented. We take
	// the more precise explanation here
	// https://github.com/EricZimmerman/KapeFiles/issues/1038 as the
	// meaning of these fields.

	// NOTE: The following fields should not be used for pure
	// Velociraptor rules, since they all boil down to simple glob -
	// use the Glob instead. We support them only to be able to
	// consume target files from the KapeFile project.
	Path      string `json:"Path,omitempty"`
	FileMask  string `json:"FileMask,omitempty"`
	Recursive bool   `json:"Recursive,omitempty"`

	// The following are ignored.
	AlwaysAddToQueue bool   `json:"AlwaysAddToQueue,omitempty"`
	SaveAsFileName   string `json:"SaveAsFileName,omitempty"`

	// Where the target file was loaded from
	file string
}

type TargetFile struct {
	Name        string `json:"Name,omitempty"`
	Description string `json:"Description,omitempty"`
	Author      string `json:"Author,omitempty"`

	// These are ignored - for compatibility with KapeFile
	RecreateDirectories bool `json:"RecreateDirectories,omitempty"`

	// This is what KapeFile calls them but this name is unclear. In
	// new target specifications, use Rules.
	Targets []*TargetRule `json:"Targets,omitempty"`

	// A list of rules to collect
	Rules []*TargetRule `json:"Rules,omitempty"`

	// Ignored
	Version string `json:"Version,omitempty"`
	Id      string `json:"Id,omitempty"`

	// If specified we copy it into the artifact export section.
	Preamble string `json:"Preamble,omitempty"`
}

type Config struct {
	Name              string   `json:"Name,omitempty"`
	Description       string   `json:"Description,omitempty"`
	TargetDirectories []string `json:"TargetDirectories,omitempty"`
	TargetRegex       string   `json:"TargetRegex,omitempty"`

	// Path to the artifact template
	ArtifactTemplate string `json:"ArtifactTemplate,omitempty"`

	RegExToGlob map[string]string `json:"RegExToGlob,omitempty"`

	// Where to store the final YAML
	Output []string `json:"Output,omitempty"`

	// Where to store the state file. The State File allows us to keep
	// track of changes in the underlying target files as the upstream
	// KapeFile repo is being tracked.
	StateFile string `json:"StateFile,omitempty"`

	// Can be / or \\
	PathSep string `json:"PathSep,omitempty"`

	Transformer string `json:"Transformer,omitempty"`

	SkipFiles []string `json:"SkipFiles"`

	// Set to build artifact in debug mode
	Debug bool `json:"Debug"`

	// Create a _All target with all the targets enabled.
	MakeAllTarget bool `json:"MakeAllTarget"`
}

type Transformer func(
	config *Config, filename string, in []byte) ([]byte, error)
