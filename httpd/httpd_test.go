package httpd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/layers"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitHTTPD(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "HTTPD", testHTTPD, spec.Report(report.Terminal{}))
}

func testHTTPD(t *testing.T, when spec.G, it spec.S) {
	when("NewContributor", func() {
		var stubHTTPDFixture = filepath.Join("fixtures", "stub-httpd.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{})
			f.AddDependency(Dependency, stubHTTPDFixture)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("should contribute httpd to launch when launch is true", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(Dependency, stubHTTPDFixture)

			nodeContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(Dependency)

			Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			Expect(layer).To(test.HaveOverrideLaunchEnvironment("APP_ROOT", f.Build.Application.Root))
			Expect(layer).To(test.HaveOverrideLaunchEnvironment("SERVER_ROOT", layer.Root))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(f.Build.Layers).To(test.HaveLaunchMetadata(
				layers.Metadata{Processes: []layers.Process{{"web", fmt.Sprintf(`httpd -f %s -k start -DFOREGROUND -C "PassEnv PORT"`, filepath.Join(f.Build.Application.Root, "httpd.conf"))}}},
			))
		})
	})
}
