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
	"github.com/paketo-buildpacks/libpak/bindings"
)

type Detect struct{}

func (Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	if _, ok, err := bindings.ResolveOne(context.Platform.Bindings, bindings.OfType("NewRelic")); err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to resolve binding NewRelic\n%w", err)
	} else if !ok {
		return libcnb.DetectResult{Pass: false}, nil
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "new-relic-java"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "new-relic-java"},
					{Name: "jvm-application"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "new-relic-nodejs"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "new-relic-nodejs"},
					{Name: "node", Metadata: map[string]interface{}{"build": true}},
					{Name: "node_modules"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "new-relic-php"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "new-relic-php"},
					{Name: "php"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "new-relic-python-config"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "new-relic-python-config"},
					{Name: "cpython", Metadata: map[string]interface{}{"build": true}},
					{Name: "site-packages", Metadata: map[string]interface{}{"build": true}},
				},
			},
		},
	}, nil
}
