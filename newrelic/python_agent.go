package newrelic

import (
	"errors"
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
	e := filepath.Walk(filepath.Join(layer.Path, "/layers/paketo-buildpacks_pip-install/packages/"), func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "newrelic") {
			return io.EOF
		}
		return nil
	})

	if e != io.EOF {
		return libcnb.Layer{}, fmt.Errorf("new relic python agent not installed, check your requirementes.txt file")
	}

	file := filepath.Join(p.ApplicationPath, "newrelic.ini")
	if _, err := os.Stat(file); err == nil {
		return layer, nil

	} else if errors.Is(err, os.ErrNotExist) {
		file := filepath.Join(p.BuildpackPath, "resources", "newrelic.ini")
		in, err := os.Open(file)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to open %s\n%w", file, err)
		}
		defer in.Close()

		file = filepath.Join(p.ApplicationPath, "newrelic.ini")
		if err := sherpa.CopyFile(in, file); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy %s to %s\n%w", in.Name(), file, err)
		}
	} else {
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install python agent\n%w", err)
		}
	}
	return layer, nil
}

func (p PythonAgent) Name() string {
	return "new-relic-python-agent"
}
