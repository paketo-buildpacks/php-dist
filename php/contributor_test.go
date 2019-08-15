/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package php

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitContributor(t *testing.T) {
	spec.Run(t, "Contributor", testContributor, spec.Report(report.Terminal{}))
}

func testContributor(t *testing.T, when spec.G, it spec.S) {
	var stubPHPFixture = filepath.Join("testdata", "stub-php.tar.gz")
	var f *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)
		f = test.NewBuildFactory(t)
	})

	it("returns true if build plan exists and version is set", func() {
		f.AddDependency(Dependency, stubPHPFixture)
		f.AddPlan(buildpackplan.Plan{
			Name:    Dependency,
			Version: "*",
		})

		_, ok, err := NewContributor(f.Build)
		Expect(ok).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	it("returns false if build plan does not exist", func() {
		_, ok, err := NewContributor(f.Build)
		Expect(ok).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())
	})

	it("contributes PHP to build", func() {
		f.AddDependency(Dependency, stubPHPFixture)
		f.AddPlan(buildpackplan.Plan{
			Name:     Dependency,
			Version:  "*",
			Metadata: buildpackplan.Metadata{"build": true},
		})

		c, shouldContribute, err := NewContributor(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(shouldContribute).To(BeTrue())

		Expect(c.Contribute()).To(Succeed())

		layer := f.Build.Layers.Layer("php-binary")
		Expect(layer).To(test.HaveLayerMetadata(true, true, false))
		Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		Expect(layer).To(test.HaveOverrideSharedEnvironment("MIBDIRS", filepath.Join(layer.Root, "mibs")))
		Expect(layer).To(test.HaveAppendPathSharedEnvironment("PATH", filepath.Join(layer.Root, "sbin")))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_HOME", filepath.Join(layer.Root)))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_EXTENSION_DIR", filepath.Join(layer.Root, "lib/php/extensions/no-debug-non-zts-12345")))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_API", "12345"))
	})

	it("contributes PHP to launch", func() {
		f.AddDependency(Dependency, stubPHPFixture)
		f.AddPlan(buildpackplan.Plan{
			Name:     Dependency,
			Version:  "*",
			Metadata: buildpackplan.Metadata{"launch": true},
		})

		c, shouldContribute, err := NewContributor(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(shouldContribute).To(BeTrue())

		Expect(c.Contribute()).To(Succeed())

		layer := f.Build.Layers.Layer("php-binary")
		Expect(layer).To(test.HaveLayerMetadata(false, false, true))
		Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		Expect(layer).To(test.HaveOverrideSharedEnvironment("MIBDIRS", filepath.Join(layer.Root, "mibs")))
		Expect(layer).To(test.HaveAppendPathSharedEnvironment("PATH", filepath.Join(layer.Root, "sbin")))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_HOME", filepath.Join(layer.Root)))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_EXTENSION_DIR", filepath.Join(layer.Root, "lib/php/extensions/no-debug-non-zts-12345")))
		Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_API", "12345"))
	})
}
