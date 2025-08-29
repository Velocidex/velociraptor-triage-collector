package testing

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/suite"
)

const (
	VelociraptorUrl        = "https://github.com/Velocidex/velociraptor/releases/download/v0.74/velociraptor-v0.74.5-linux-amd64-musl"
	VelociraptorBinaryPath = "./velociraptor.bin"
)

type TestSuite struct {
	suite.Suite
}

func (self *TestSuite) SetupSuite() {
	self.findAndPrepareBinary()
}

func (self *TestSuite) findAndPrepareBinary() {
	t := self.T()

	_, err := os.Lstat(VelociraptorBinaryPath)
	if err != nil {
		fmt.Printf("Downloading %v from %v\n", VelociraptorBinaryPath,
			VelociraptorUrl)
		resp, err := http.Get(VelociraptorUrl)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Create the file
		out, err := os.OpenFile(VelociraptorBinaryPath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
		assert.NoError(t, err)
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		assert.NoError(t, err)
	}
}

func (self *TestSuite) TestGolden() {
	t := self.T()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	test_files, err := filepath.Abs(filepath.Join(cwd, "../test_files/"))
	assert.NoError(t, err)

	config_file, err := filepath.Abs(filepath.Join(cwd, "test.config.yaml"))
	assert.NoError(t, err)

	test_cases, err := filepath.Abs(filepath.Join(cwd, "./testcases/"))
	assert.NoError(t, err)

	argv := []string{
		"--definitions", "../output", "-v",
		"--config", config_file,
		"golden", "--env", "testFiles=" + test_files, test_cases,
	}

	out, err := runWithArgs(argv)
	assert.NoError(t, err, string(out))

	fmt.Println(string(out))
}

func TestArtifact(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func filterOut(out string) []string {
	res := []string{}

	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "OSPath") {
			res = append(res, line)
		}
	}
	return res
}

func runWithArgs(argv []string, args ...string) (string, error) {
	full_argv := append(argv, args...)
	os.Setenv("VELOCIRAPTOR_CONFIG", "")

	fmt.Printf("Running %v %v\n", VelociraptorBinaryPath,
		strings.Join(full_argv, " "))
	cmd := exec.Command(VelociraptorBinaryPath, full_argv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
