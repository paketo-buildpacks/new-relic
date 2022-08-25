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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type PHPAgent struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewPHPAgent(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) (PHPAgent, libcnb.BOMEntry) {
	contributor, entry := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Launch: true,
	})
	return PHPAgent{LayerContributor: contributor}, entry
}

func (p PHPAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	p.LayerContributor.Logger = p.Logger

	return p.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		p.Logger.Bodyf("Expanding to %s", layer.Path)

		if err := crush.ExtractTarGz(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand New Relic\n%w", err)
		}

		s := fmt.Sprintf(`extension = %[1]s/agent/x64/newrelic-${PHP_API}.so

[newrelic]
newrelic.appname = ${NEW_RELIC_APP_NAME}
newrelic.license = ${NEW_RELIC_LICENSE_KEY}
newrelic.logfile = /proc/self/fd/1
newrelic.daemon.logfile = %[1]s/newrelic-daemon.log
newrelic.daemon.location = %[1]s/daemon/newrelic-daemon.x64
`, layer.Path)

		file := filepath.Join(layer.Path, "php.ini.d")
		if err := os.MkdirAll(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to create %s\n%w", file, err)
		}

		layer.LaunchEnvironment.Prepend("PHP_INI_SCAN_DIR", string(os.PathListSeparator), file)

		file = filepath.Join(layer.Path, "php.ini.d", "newrelic.ini")
		if err := ioutil.WriteFile(file, []byte(s), 0644); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to write %s\n%w", file, err)
		}

		return layer, nil
	})
}

func (p PHPAgent) Name() string {
	return p.LayerContributor.LayerName()
}
