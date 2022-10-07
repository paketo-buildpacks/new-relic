package newrelic

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type PythonAgent struct {
	BuildpackPath   string
	ApplicationPath string
	Logger          bard.Logger
}

func NewPythonAgent(applicationPath string, buildpackPath string) PythonAgent {
	return PythonAgent{
		ApplicationPath: applicationPath,
		BuildpackPath:   buildpackPath,
	}
}

func (p PythonAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	e := filepath.Walk("/layers/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		if err == nil && strings.Contains(path, "/newrelic/admin") {
			return io.EOF
		}
		return nil
	})
	if e != io.EOF {
		return libcnb.Layer{}, fmt.Errorf("new relic python agent not installed, check your requirementes.txt file")
	}
	config_file := filepath.Join(p.ApplicationPath, "newrelic.ini")
	file_to_copy := filepath.Join(p.BuildpackPath, "resources", "newrelic.ini")
	if _, err := sherpa.FileExists(config_file); err != nil {
		return layer, nil

	} else {
		if _, err := sherpa.FileExists(file_to_copy); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to open %s\n%w", file_to_copy, err)
		} else {
			in, _ := os.Open(file_to_copy)
			if err := sherpa.CopyFile(in, config_file); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to copy newrelic.ini to configure New Relic Python agent\n%w", err)
			}
		}
	}
	return layer, nil
}

func (p PythonAgent) Name() string {
	return "new-relic-python-config"
}
