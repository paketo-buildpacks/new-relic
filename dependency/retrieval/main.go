// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/buildpack_config"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	javaLicenseURI = "https://newrelic.com/docs/java/java-agent-license"
	licenseURI     = "https://docs.newrelic.com/docs/licenses/license-information"

	javaGroupID     = "com.newrelic.agent.java"
	javaArtifactID  = "newrelic-agent"
	mavenRepository = "https://repo1.maven.org/maven2"

	phpArchive  = "https://download.newrelic.com/php_agent/archive"
	dotnetIndex = "https://download.newrelic.com/dot_net_agent/previous_releases/"
)

var (
	javaVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	phpLinkPattern     = regexp.MustCompile(`/php_agent/archive/(\d+)\.(\d+)\.(\d+)\.(\d+)`)
	dotnetLinkPattern  = regexp.MustCompile(`/dot_net_agent/previous_releases/(\d+)\.(\d+)\.(\d+)`)
	buildPattern       = regexp.MustCompile(`(\d+)$`)
	javaArtifactBase   = fmt.Sprintf("%s/%s/%s", mavenRepository, strings.ReplaceAll(javaGroupID, ".", "/"), javaArtifactID)
)

type dependency struct {
	id         string
	name       string
	purl       string
	cpeVendor  string
	cpeProduct string
	arch       string
	licenseURI string
	getAll     retrieve.GetAllVersionsFunc
}

type version struct {
	version *semver.Version
	build   string
	uri     string
	source  string
}

func (v version) Version() *semver.Version {
	return v.version
}

var dependencies = []dependency{
	{
		id:         "new-relic-java",
		name:       "New Relic Java Agent",
		purl:       "pkg:generic/newrelic-java-agent",
		cpeVendor:  "newrelic",
		cpeProduct: "java-agent",
		licenseURI: javaLicenseURI,
		getAll:     getAllJavaVersions,
	},
	{
		id:         "new-relic-php",
		name:       "New Relic PHP Agent",
		purl:       "pkg:generic/newrelic-php-agent",
		cpeVendor:  "appdynamics",
		cpeProduct: "php-agent",
		arch:       "amd64",
		licenseURI: licenseURI,
		getAll:     getAllPHPVersions,
	},
	{
		id:         "new-relic-dotnet",
		name:       "New Relic Dotnet Agent",
		purl:       "pkg:generic/newrelic-dotnet-agent",
		cpeVendor:  "appdynamics",
		cpeProduct: "dotnet-agent",
		arch:       "amd64",
		licenseURI: licenseURI,
		getAll:     getAllDotnetVersions,
	},
}

type mavenMetadata struct {
	Versioning struct {
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

func main() {
	buildpackTomlPath, output := retrieve.FetchArgs()

	config, err := buildpack_config.ParseBuildpackToml(buildpackTomlPath)
	if err != nil {
		panic(err)
	}

	var metadata []versionology.Dependency
	for _, dep := range dependencies {
		newVersions, err := retrieve.GetNewVersionsForId(dep.id, config, dep.getAll)
		if err != nil {
			panic(err)
		}

		metadata = append(metadata, retrieve.GenerateAllMetadata(newVersions, generateMetadata(dep))...)
	}

	content, err := json.Marshal(metadata)
	if err != nil {
		panic(fmt.Errorf("unable to marshal metadata\n%w", err))
	}

	if err := os.WriteFile(output, content, 0o644); err != nil {
		panic(fmt.Errorf("cannot write to %s: %w", output, err))
	}
	fmt.Printf("Wrote metadata to %s\n", output)
}

func getAllJavaVersions() (versionology.VersionFetcherArray, error) {
	metadataURL := fmt.Sprintf("%s/maven-metadata.xml", javaArtifactBase)

	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch %s\n%w", metadataURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch %s: status code %d", metadataURL, resp.StatusCode)
	}

	var metadata mavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("unable to parse %s\n%w", metadataURL, err)
	}

	var versions versionology.VersionFetcherArray
	for _, versionString := range metadata.Versioning.Versions.Version {
		if !javaVersionPattern.MatchString(versionString) {
			continue
		}

		v, err := semver.NewVersion(versionString)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", versionString)
			continue
		}

		versions = append(versions, version{version: v, build: "0"})
	}

	return versions, nil
}

func getAllPHPVersions() (versionology.VersionFetcherArray, error) {
	body, err := fetchBody(fmt.Sprintf("%s/index.html", phpArchive))
	if err != nil {
		return nil, err
	}

	latest := map[string]version{}
	for _, m := range phpLinkPattern.FindAllStringSubmatch(body, -1) {
		full := fmt.Sprintf("%s.%s.%s.%s", m[1], m[2], m[3], m[4])
		versionString := fmt.Sprintf("%s.%s.%s", m[1], m[2], m[3])

		v, err := semver.NewVersion(versionString)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", versionString)
			continue
		}

		candidate := version{
			version: v,
			build:   m[4],
			uri:     fmt.Sprintf("%s/%s/newrelic-php5-%s-linux.tar.gz", phpArchive, full, full),
			source:  fmt.Sprintf("https://github.com/newrelic/newrelic-php-agent/archive/refs/tags/v%s.tar.gz", full),
		}

		if existing, ok := latest[versionString]; !ok || buildLess(existing.build, candidate.build) {
			latest[versionString] = candidate
		}
	}

	return toArray(latest), nil
}

func getAllDotnetVersions() (versionology.VersionFetcherArray, error) {
	body, err := fetchBody(dotnetIndex)
	if err != nil {
		return nil, err
	}

	latest := map[string]version{}
	for _, m := range dotnetLinkPattern.FindAllStringSubmatch(body, -1) {
		versionString := fmt.Sprintf("%s.%s.%s", m[1], m[2], m[3])

		v, err := semver.NewVersion(versionString)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", versionString)
			continue
		}

		candidate := version{
			version: v,
			build:   "0",
			uri: fmt.Sprintf("%s%s/newrelic-dotnet-agent_%s_amd64.tar.gz",
				dotnetIndex, versionString, versionString),
			source: fmt.Sprintf("https://github.com/newrelic/newrelic-dotnet-agent/archive/refs/tags/v%s.tar.gz", versionString),
		}

		if _, ok := latest[versionString]; !ok {
			latest[versionString] = candidate
		}
	}

	return toArray(latest), nil
}

func toArray(latest map[string]version) versionology.VersionFetcherArray {
	var versions versionology.VersionFetcherArray
	for _, v := range latest {
		versions = append(versions, v)
	}

	return versions
}

func generateMetadata(dep dependency) retrieve.GenerateMetadataFunc {
	return func(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
		v, ok := versionFetcher.(version)
		if !ok {
			return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
		}

		versionString := v.version.String()

		uri := v.uri
		source := v.source
		if uri == "" {
			uri = fmt.Sprintf("%s/%s/%s-%s.jar", javaArtifactBase, versionString, javaArtifactID, versionString)
			source = fmt.Sprintf("%s/%s/%s-%s-sources.jar", javaArtifactBase, versionString, javaArtifactID, versionString)
		}

		checksum, err := upstream.GetSHA256OfRemoteFile(uri)
		if err != nil {
			return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
		}

		sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
		if err != nil {
			return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
		}

		purl := fmt.Sprintf("%s@%s", dep.purl, versionString)
		if dep.arch != "" {
			purl = fmt.Sprintf("%s?arch=%s", purl, dep.arch)
		}

		dependency := cargo.ConfigMetadataDependency{
			Checksum: fmt.Sprintf("sha256:%s", checksum),
			CPE:      fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*", dep.cpeVendor, dep.cpeProduct, versionString),
			ID:       dep.id,
			Licenses: []interface{}{
				map[string]string{
					"uri": dep.licenseURI,
				},
			},
			Name:           dep.name,
			PURL:           purl,
			SHA256:         checksum,
			Source:         source,
			SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
			Stacks:         []string{"*"},
			URI:            uri,
			Version:        versionString,
		}

		return versionology.NewDependencyArray(dependency, "")
	}
}

func fetchBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("unable to fetch %s\n%w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to fetch %s: status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read %s\n%w", url, err)
	}

	return string(body), nil
}

// buildLess reports whether build a is older than build b. Builds are numeric
// strings; compare them as integers with a string fallback.
func buildLess(a, b string) bool {
	an := buildPattern.FindString(a)
	bn := buildPattern.FindString(b)
	if an == "" || bn == "" {
		return a < b
	}

	ai, err := strconv.Atoi(an)
	if err != nil {
		return a < b
	}
	bi, err := strconv.Atoi(bn)
	if err != nil {
		return a < b
	}

	return ai < bi
}
