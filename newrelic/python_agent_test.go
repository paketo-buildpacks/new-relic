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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"

	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/effect/mocks"
	"github.com/paketo-buildpacks/new-relic/v4/newrelic"
)

func testPythonAgent(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx      libcnb.BuildContext
		executor *mocks.Executor
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "python-agent-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Buildpack.Path, err = ioutil.TempDir("", "python-agent-buildpack")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "python-agent-layers")
		Expect(err).NotTo(HaveOccurred())

		executor = &mocks.Executor{}
		executor.On("Execute", mock.Anything).Return(nil)
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Buildpack.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes Python agent", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "resources"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "resources", "newrelic.ini"), []byte{}, 0644)).
			To(Succeed())

		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-new-relic-agent.tgz",
			SHA256: "e6417c651cc4d3fbc0ece8c715f8098106cda1a19036805fa4746db9f05b2e9a",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		p, _ := newrelic.NewPythonAgent(ctx.Application.Path, ctx.Buildpack.Path, dep, dc)
		p.Executor = executor
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = p.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		execution := executor.Calls[0].Arguments[0].(effect.Execution)
		Expect(execution.Command).To(Equal("pip3"))
		Expect(execution.Args).To(Equal([]string{"install", "newrelic==" + os.Getenv("NEWRELIC_PYTHON_VERSION")}))
		Expect(layer.LaunchEnvironment["NEWRELIC_PYTHON_VERSION"]).To(Equal(os.Getenv("NEWRELIC_PYTHON_VERSION")))
		Expect(filepath.Join(p.ApplicationPath, "newrelic.ini")).To(BeARegularFile())
	})
}
