// MIT License
//
// Copyright (c) 2016 Nick Miyake
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gofiles_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/nmiyake/pkg/gofiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteGoFiles(t *testing.T) {
	restoreEnvVars := setEnvVars(map[string]string{
		"GOFLAGS": "",
	})
	defer restoreEnvVars()

	dir, cleanup, err := dirs.TempDir("", "")
	require.NoError(t, err)
	defer cleanup()

	goFiles, err := gofiles.Write(dir, []gofiles.GoFileSpec{
		{
			RelPath: "go.mod",
			Src: `module github.com/foo

require github.com/baz v1.0.0

replace github.com/baz => ./baz
`,
		},
		{
			RelPath: "foo.go",
			Src: `package main
import (
	"fmt"
	"github.com/foo/bar"
	"github.com/baz"
)
func main() {
	fmt.Println(bar.Bar(), baz.Baz())
}`,
		},
		{
			RelPath: "bar/bar.go",
			Src: `package bar
func Bar() string {
	return "bar"
}`,
		},
		{
			RelPath: "baz/go.mod",
			Src:     `module github.com/baz`,
		},
		{
			RelPath: "baz/baz.go",
			Src: `package baz
func Baz() string {
	return "baz"
}`,
		},
	})
	require.NoError(t, err)

	goRunCmd := exec.Command("go", "run", "foo.go")
	goRunCmd.Dir = filepath.Dir(goFiles["foo.go"].Path)
	output, err := goRunCmd.CombinedOutput()
	require.NoError(t, err)

	assert.Equal(t, "bar baz\n", string(output))
}

func setEnvVars(envVars map[string]string) func() {
	origVars := make(map[string]string)
	var unsetVars []string
	for k := range envVars {
		val, ok := os.LookupEnv(k)
		if !ok {
			unsetVars = append(unsetVars, k)
			continue
		}
		origVars[k] = val
	}

	for k, v := range envVars {
		_ = os.Setenv(k, v)
	}

	return func() {
		for _, k := range unsetVars {
			_ = os.Unsetenv(k)
		}
		for k, v := range origVars {
			_ = os.Setenv(k, v)
		}
	}
}
