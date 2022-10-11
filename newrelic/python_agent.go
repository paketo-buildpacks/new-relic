/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
 
package newrelic

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type PythonAgent struct {
	BuildpackPath   string
	ApplicationPath string
	Executor        effect.Executor
	Logger          bard.Logger
}

func NewPythonAgent(applicationPath string, buildpackPath string, executor effect.Executor) PythonAgent {
	return PythonAgent{
		ApplicationPath: applicationPath,
		BuildpackPath:   buildpackPath,
		Executor:        executor,
	}
}

func (p PythonAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	err := p.Executor.Execute(effect.Execution{
		Command: "python",
		Args:    []string{"-c", "import newrelic.agent"},
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to run python\n%w", err)
	}

	config_file := filepath.Join(p.ApplicationPath, "newrelic.ini")
	file_to_copy := filepath.Join(p.BuildpackPath, "resources", "newrelic.ini")
	if found, err := sherpa.FileExists(config_file); err != nil && found {
		return layer, nil
	} else {
		if _, err := sherpa.FileExists(file_to_copy); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to open %s\n%w", file_to_copy, err)
		} else {
			in, err := os.Open(file_to_copy)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to open newrelic.ini\n%w", err)
			}

			if err := sherpa.CopyFile(in, config_file); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to copy newrelic.ini\n%w", err)
			}
		}
	}
	return layer, nil
}

func (p PythonAgent) Name() string {
	return "new-relic-python-config"
}
