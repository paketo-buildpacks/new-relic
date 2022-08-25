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
	"strconv"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type Extensions struct {
	ConfigurationResolver libpak.ConfigurationResolver
	LayerContributor      libpak.DependencyLayerContributor
	Logger                bard.Logger
}

func NewExtensions(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, configurationResolver libpak.ConfigurationResolver) (Extensions, libcnb.BOMEntry) {
	contributor, entry := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Launch: true,
	})
	return Extensions{
		ConfigurationResolver: configurationResolver,
		LayerContributor:      contributor,
	}, entry
}

func (e Extensions) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	e.LayerContributor.Logger = e.Logger

	return e.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		e.Logger.Bodyf("Expanding to %s", layer.Path)

		var err error
		c := 0
		if s, ok := e.ConfigurationResolver.Resolve("BP_NEW_RELIC_EXT_STRIP"); ok {
			if c, err = strconv.Atoi(s); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to parse %s to integer\n%w", s, err)
			}
		}

		if err := crush.ExtractTarGz(artifact, layer.Path, c); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand extensions\n%w", err)
		}

		layer.LaunchEnvironment.Appendf("JAVA_TOOL_OPTIONS", " ", "-Dnewrelic.config.extensions.dir=%s", layer.Path)

		return layer, nil
	})
}

func (e Extensions) Name() string {
	return e.LayerContributor.LayerName()
}
