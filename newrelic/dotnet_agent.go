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
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type DotnetAgent struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewDotnetAgent(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) (DotnetAgent, libcnb.BOMEntry) {
	contributor, entry := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Launch: true,
	})
	return DotnetAgent{
		LayerContributor: contributor,
	}, entry
}

func (d DotnetAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	d.LayerContributor.Logger = d.Logger

	layer, err := d.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		d.Logger.Bodyf("Expanding to %s", layer.Path)

		if err := crush.ExtractTarGz(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand New Relic\n%w", err)
		}

		layer.LaunchEnvironment.Default("CORECLR_NEWRELIC_HOME", layer.Path)
		layer.LaunchEnvironment.Default("CORECLR_ENABLE_PROFILING", "1")
		layer.LaunchEnvironment.Default("CORECLR_PROFILER", "{36032161-FFC0-4B61-B559-F6C5D41BAE5A}")
		layer.LaunchEnvironment.Default("CORECLR_PROFILER_PATH", filepath.Join(layer.Path, "libNewRelicProfiler.so"))

		return layer, nil
	})
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to contribute dotnet agent\n%w", err)
	}

	return layer, nil
}

func (n DotnetAgent) Name() string {
	return n.LayerContributor.LayerName()
}
