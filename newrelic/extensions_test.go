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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/new-relic/v4/newrelic"
)

func testExtensions(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "properties-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes extensions", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-extensions.tar.gz",
			SHA256: "22e708cfd301430cbcf8d1c2289503d8288d50df519ff4db7cca0ff9fe83c324",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		p, _ := newrelic.NewExtensions(dep, dc, libpak.ConfigurationResolver{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = p.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
		Expect(layer.LaunchEnvironment["JAVA_TOOL_OPTIONS.delim"]).To(Equal(" "))
		Expect(layer.LaunchEnvironment["JAVA_TOOL_OPTIONS.append"]).To(
			Equal(fmt.Sprintf("-Dnewrelic.config.extensions.dir=%s", layer.Path)))
	})

	context("$BP_NEW_RELIC_EXT_STRIP", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_NEW_RELIC_EXT_STRIP", "1")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_NEW_RELIC_EXT_STRIP")).To(Succeed())
		})

		it("contributes extensions with directory", func() {
			dep := libpak.BuildpackDependency{
				URI:    "https://localhost/stub-extensions-with-directory.tar.gz",
				SHA256: "3d3e33f59551b6df159c772bf755d824668019dad345d9401100245c12de4b9b",
			}
			dc := libpak.DependencyCache{CachePath: "testdata"}

			p, _ := newrelic.NewExtensions(dep, dc, libpak.ConfigurationResolver{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = p.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
			Expect(layer.LaunchEnvironment["JAVA_TOOL_OPTIONS.delim"]).To(Equal(" "))
			Expect(layer.LaunchEnvironment["JAVA_TOOL_OPTIONS.append"]).To(
				Equal(fmt.Sprintf("-Dnewrelic.config.extensions.dir=%s", layer.Path)))
		})
	})
}
