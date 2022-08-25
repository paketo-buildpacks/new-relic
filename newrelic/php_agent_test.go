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

func testPHPAgent(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "php-agent-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes PHP agent", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-new-relic-agent.tar.gz",
			SHA256: "52d21d0eac639a8b5d47f52b0256a5acb94983282ecfa8f7a11bbaf85da77425",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j, _ := newrelic.NewPHPAgent(dep, dc)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
		Expect(ioutil.ReadFile(filepath.Join(layer.Path, "php.ini.d", "newrelic.ini"))).To(Equal([]byte(fmt.Sprintf(`extension = %[1]s/agent/x64/newrelic-${PHP_API}.so

[newrelic]
newrelic.appname = ${NEW_RELIC_APP_NAME}
newrelic.license = ${NEW_RELIC_LICENSE_KEY}
newrelic.logfile = /proc/self/fd/1
newrelic.daemon.logfile = %[1]s/newrelic-daemon.log
newrelic.daemon.location = %[1]s/daemon/newrelic-daemon.x64
`, layer.Path))))
		Expect(layer.LaunchEnvironment["PHP_INI_SCAN_DIR.delim"]).To(Equal(string(os.PathListSeparator)))
		Expect(layer.LaunchEnvironment["PHP_INI_SCAN_DIR.prepend"]).To(Equal(filepath.Join(layer.Path, "php.ini.d")))
	})
}
