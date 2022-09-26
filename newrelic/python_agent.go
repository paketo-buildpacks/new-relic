package newrelic

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type PythonAgent struct {
	BuildpackPath    string
	ApplicationPath  string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewPythonAgent(applicationPath string, buildpackPath string) PythonAgent {
	return PythonAgent{
		ApplicationPath: applicationPath,
		BuildpackPath:   buildpackPath,
	}
}

func (p PythonAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	p.LayerContributor.Logger = p.Logger

	p.Logger.Bodyf("Installing to %s", layer.Path)

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

	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to install python agent\n%w", err)
	}
	return layer, nil
}

func (p PythonAgent) Name() string {
	return p.LayerContributor.LayerName()
}
