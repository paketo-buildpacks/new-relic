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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	if _, ok, err := pr.Resolve("new-relic-java"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve new-relic-java plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("new-relic-java", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		ja, be := NewJavaAgent(context.Buildpack.Path, dep, dc)
		ja.Logger = b.Logger
		result.Layers = append(result.Layers, ja)
		result.BOM.Entries = append(result.BOM.Entries, be)

		if uri, ok := cr.Resolve("BP_NEW_RELIC_EXT_URI"); ok {
			v, _ := cr.Resolve("BP_NEW_RELIC_EXT_VERSION")
			s, _ := cr.Resolve("BP_NEW_RELIC_EXT_SHA256")

			dep = libpak.BuildpackDependency{
				ID:      "new-relic-extensions",
				Name:    "New Relic Extensions",
				Version: v,
				URI:     uri,
				SHA256:  s,
				Stacks:  []string{context.StackID},
				CPEs:    []string{fmt.Sprintf("cpe:2.3:a:newrelic:external-configuration:%s:*:*:*:*:*:*:*", v)},
				PURL:    fmt.Sprintf("pkg:generic/newrelic-external-configuration@%s", v),
			}

			ec, be := NewExtensions(dep, dc, cr)
			ec.Logger = b.Logger
			result.Layers = append(result.Layers, ec)
			result.BOM.Entries = append(result.BOM.Entries, be)
		}
	}

	if _, ok, err := pr.Resolve("new-relic-python"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve new-relic-python plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("new-relic-python", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}
		p, _ := NewPythonAgent(context.Application.Path, context.Buildpack.Path, dep, dc)
		p.Logger = b.Logger
		result.Layers = append(result.Layers, p)
	}

	if _, ok, err := pr.Resolve("new-relic-nodejs"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve new-relic-nodejs plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("new-relic-nodejs", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		na, be := NewNodeJSAgent(context.Application.Path, context.Buildpack.Path, dep, dc)
		na.Logger = b.Logger
		result.Layers = append(result.Layers, na)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	if _, ok, err := pr.Resolve("new-relic-php"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve new-relic-php plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("new-relic-php", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		pa, be := NewPHPAgent(dep, dc)
		pa.Logger = b.Logger
		result.Layers = append(result.Layers, pa)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	h, be := libpak.NewHelperLayer(context.Buildpack, "properties")
	h.Logger = b.Logger
	result.Layers = append(result.Layers, h)
	result.BOM.Entries = append(result.BOM.Entries, be)

	return result, nil
}
