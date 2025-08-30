package compiler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Velocidex/velociraptor-triage-collector/api"
)

var (
	UnknownRegexFileMask = errors.New("Unknown Regex FileMask")
	StripDriveRegex      = regexp.MustCompile("^[a-zA-Z]:(\\\\|/)")
)

// Kape targets sometimes have a regex instead of a glob - it is not
// trivial to convert a regex to a glob automatically. A regex is not
// generally necessary and just makes life complicated, so we just hard
// code these translations manually and alert when a new regex pops up.
func (self *Compiler) unregexify(regex string) (string, error) {
	if self.config_obj.RegExToGlob == nil {
		return regex, fmt.Errorf("%w: %v", UnknownRegexFileMask, regex)
	}

	res, pres := self.config_obj.RegExToGlob[regex]
	if !pres {
		return regex, fmt.Errorf("%w: %v", UnknownRegexFileMask, regex)
	}

	return res, nil
}

// Kape Target rules add some undocumented expansions that dont mean
// anything and can not really be expanded in runtime - we just
// replace them with * glob.
func (self *Compiler) remove_fluff(glob string) string {
	for _, v := range []string{
		"%user%", "%users%", "%User%", "%Users%"} {
		glob = strings.ReplaceAll(glob, v, "*")
	}
	return glob

}

// Clears all the lagacy fields from KapeFile format.
func (self *Compiler) clearLegacyRule(t *api.TargetRule) {
	t.Path = ""
	t.FileMask = ""
	t.Recursive = false
	t.AlwaysAddToQueue = false
	t.SaveAsFileName = ""
}

// Parse the target and ensure it is valid.
func (self *Compiler) ValidateRule(
	t *api.TargetRule, target_file *api.TargetFile) (err error) {
	// Handle old KapeFile style rule definitions.
	if t.Glob == "" && t.Path != "" {
		// When we are done clean up all the lagacy fields.
		defer self.clearLegacyRule(t)

		// Convert the target description to a glob. This is done using a
		// complicated interactions between various attributes of the
		// target description that are not that well documented. However,
		// they were clarified by the KapeFiles maintainers here
		// https://github.com/EricZimmerman/KapeFiles/issues/1038

		// Ultimately it all boils down to a simple glob expression,
		// so we recreate the glob expression and wipe the other
		// fields.

		// The Path always represents a directory
		base_glob := t.Path

		// Actually refering to other targets. We need to resolve it
		// at runtime.
		if strings.HasSuffix(base_glob, ".tkape") {
			t.Ref = strings.TrimSuffix(base_glob, ".tkape")
			return nil
		}

		recursive := ""
		if t.Recursive {
			recursive = self.config_obj.PathSep + "**"
		}

		// The default FileMask is *
		mask := "*"
		if t.FileMask != "" {
			mask = t.FileMask
		}

		if strings.HasPrefix(strings.ToLower(mask), "regex:") {
			mask, err = self.unregexify(mask[6:])
			if err != nil {
				return err
			}
		}
		mask = self.config_obj.PathSep + mask

		// To simplify the glob reduce suffix of /**10/* to just
		// /**10
		if recursive != "" && mask == self.config_obj.PathSep+"*" {
			mask = ""
		}

		t.Glob = strings.TrimSuffix(base_glob, self.config_obj.PathSep) +
			recursive + mask
		t.Glob = StripDriveRegex.ReplaceAllString(t.Glob, "")
		t.Glob = self.remove_fluff(t.Glob)

		/*
		   # If Recursive is specified, it means we recurse into the directory.
		   recursive = ""
		   if target.get("Recursive") or ctx.kape_data.get("RecreateDirectories"):
		       recursive = "/**10"

		   # The default FileMask is *
		   mask = target.get("FileMask", "*")
		   if mask.lower().startswith("regex:"):
		       mask = unregexify(mask[6:])

		   mask = "/" + mask

		   # To simplify the glob reduce suffix of /**10/* to just
		   # /**10
		   if recursive and mask == "/*":
		       mask = ""

		   glob = base_glob.rstrip("\\") + recursive + mask

		   row_id = ctx.resolve_id(name, glob)
		   ctx.groups[name].add(row_id)

		   glob = strip_drive(glob)
		   glob = remove_fluff(glob)
		   glob = ctx.pathsep_converter(glob)

		*/
	}
	return nil
}
