package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo create binding command tests", func() {
	var commonVar helper.CommonVar
	var devSession helper.DevSession
	var err error

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		createdBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(createdBindableKind.ExitCode()).To(BeEquivalentTo(0))
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("the component is created", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
		})
		When("creating a binding", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "add", "binding", "--name", "my-binding", "--service", "cluster-sample").ShouldPass()
			})
			It("should successfully add binding between component and service in the devfile", func() {
				components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), "my-binding")
				Expect(components).ToNot(BeNil())
			})
			When("odo dev is run", func() {
				BeforeEach(func() {
					devSession, _, _, _, err = helper.StartDevMode()
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				It("should successfully bind component and service", func() {
					stdout := commonVar.CliRunner.Run("get", "servicebinding", "my-binding").Out.Contents()
					Expect(stdout).To(ContainSubstring("ApplicationsBound"))
				})
				When("odo dev command is stopped", func() {
					BeforeEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should have successfully delete the binding", func() {
						_, errOut := commonVar.CliRunner.GetServiceBinding("my-binding", commonVar.Project)
						Expect(errOut).To(ContainSubstring("not found"))
					})
				})
			})
		})
	})
})
