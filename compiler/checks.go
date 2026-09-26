package compiler

import (
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
)

var (
	RegExChars = regexp.MustCompile(`[?()|]`)
)

func (self *Compiler) CheckConfig() error {
	for regex, glob_exp := range self.config_obj.RegExToGlob {

		// Kape seems to accept a leading * here - maybe it is a
		// special case.
		_, err := regexp.Compile(`.` + regex)
		if err != nil {
			return fmt.Errorf("RegExToGlob: %v->%v %w",
				regex, glob_exp, err)
		}

		_, err = glob.Compile(glob_exp)
		if err != nil {
			return fmt.Errorf("RegExToGlob: %v->%v %w",
				regex, glob_exp, err)
		}

		// Although these regex characters are literals in glob they
		// are unlikely to exist in the glob since this started off
		// being a regex.
		if RegExChars.MatchString(glob_exp) {
			return fmt.Errorf("RegExToGlob: possible regex chars in glob %v", glob_exp)
		}
	}
	return nil
}
