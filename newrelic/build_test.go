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

package newrelic_test

import (
	"os"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/new-relic/v4/newrelic"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it("contributes Java agent API <= 0.6", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-java"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-java",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.6"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-java"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-java"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	it("contributes Java agent API >= 0.7", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-java"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-java",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
					"cpes":    []interface{}{"cpe:2.3:a:newrelic:java-agent:1.1.1:*:*:*:*:*:*:*"},
					"purl":    "pkg:generic/newrelic-java-agent@1.1.1?arch=amd64",
				},
			},
		}
		ctx.Buildpack.API = "0.7"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-java"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-java"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	context("$BP_NEW_RELIC_EXT_URI", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_NEW_RELIC_EXT_SHA256", "test-sha256")).To(Succeed())
			Expect(os.Setenv("BP_NEW_RELIC_EXT_URI", "test-uri")).To(Succeed())
			Expect(os.Setenv("BP_NEW_RELIC_EXT_VERSION", "test-version")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_NEW_RELIC_EXT_SHA256")).To(Succeed())
			Expect(os.Unsetenv("BP_NEW_RELIC_EXT_URI")).To(Succeed())
			Expect(os.Unsetenv("BP_NEW_RELIC_EXT_VERSION")).To(Succeed())
		})

		it("contributes extensions when $BP_NEW_RELIC_EXT_URI is set API <= 0.6", func() {
			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-java"})
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "new-relic-java",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
					},
				},
			}
			ctx.Buildpack.API = "0.6"
			ctx.StackID = "test-stack-id"

			result, err := newrelic.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[1].(newrelic.Extensions).LayerContributor.Dependency).To(Equal(libpak.BuildpackDependency{
				ID:      "new-relic-extensions",
				Name:    "New Relic Extensions",
				Version: "test-version",
				URI:     "test-uri",
				SHA256:  "test-sha256",
				Stacks:  []string{ctx.StackID},
				CPEs:    []string{"cpe:2.3:a:newrelic:external-configuration:test-version:*:*:*:*:*:*:*"},
				PURL:    "pkg:generic/newrelic-external-configuration@test-version",
			}))

			Expect(result.BOM.Entries).To(HaveLen(3))
			Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-java"))
			Expect(result.BOM.Entries[1].Name).To(Equal("new-relic-extensions"))
			Expect(result.BOM.Entries[2].Name).To(Equal("helper"))
		})

		it("contributes extensions when $BP_NEW_RELIC_EXT_URI is set API >= 0.7", func() {
			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-java"})
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "new-relic-java",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
						"cpes":    []interface{}{"cpe:2.3:a:newrelic:external-configuration:1.1.1:*:*:*:*:*:*:*"},
						"purl":    "pkg:generic/newrelic-external-configuration@1.1.1",
					},
				},
			}
			ctx.Buildpack.API = "0.7"
			ctx.StackID = "test-stack-id"

			result, err := newrelic.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[1].(newrelic.Extensions).LayerContributor.Dependency).To(Equal(libpak.BuildpackDependency{
				ID:      "new-relic-extensions",
				Name:    "New Relic Extensions",
				Version: "test-version",
				URI:     "test-uri",
				SHA256:  "test-sha256",
				Stacks:  []string{ctx.StackID},
				CPEs:    []string{"cpe:2.3:a:newrelic:external-configuration:test-version:*:*:*:*:*:*:*"},
				PURL:    "pkg:generic/newrelic-external-configuration@test-version",
			}))

			Expect(result.BOM.Entries).To(HaveLen(3))
			Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-java"))
			Expect(result.BOM.Entries[1].Name).To(Equal("new-relic-extensions"))
			Expect(result.BOM.Entries[2].Name).To(Equal("helper"))
		})
	})

	it("contributes NodeJS agent API <= 0.6", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-nodejs"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-nodejs",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.6"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-nodejs"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-nodejs"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	it("contributes NodeJS agent API >= 0.7", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-nodejs"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-nodejs",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
					"cpes":    []interface{}{"cpe:2.3:a:new-relic:nodejs-agent:1.1.1:*:*:*:*:*:*:*"},
					"purl":    "pkg:generic/new-relic-nodejs-agent@1.1.1?arch=amd64",
				},
			},
		}
		ctx.Buildpack.API = "0.7"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-nodejs"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-nodejs"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	it("contributes PHP agent API <= 0.6", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-php"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-php",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.6"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-php"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-php"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	it("contributes PHP agent API >= 0.7", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-php"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-php",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
					"cpes":    []interface{}{"cpe:2.3:a:newrelic:php-agent:1.1.1:*:*:*:*:*:*:*"},
					"purl":    "pkg:generic/newrelic-php-agent@1.1.1?arch=amd64",
				},
			},
		}
		ctx.Buildpack.API = "0.7"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-php"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("new-relic-php"))
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
	})

	it("contributes Python agent API >= 0.7", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "new-relic-python"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "new-relic-python",
					"version": "8.1.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.7"
		ctx.StackID = "test-stack-id"

		result, err := newrelic.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("new-relic-python"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))
	})
}
