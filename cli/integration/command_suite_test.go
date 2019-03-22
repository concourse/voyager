package integration_test

import (
	"os"
	"time"

	"github.com/onsi/gomega/gexec"

	"github.com/concourse/voyager/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"

	"testing"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Suite")
}

var postgresRunner helpers.Runner
var dbProcess ifrit.Process
var cliPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("github.com/concourse/voyager/cli")
	Expect(err).ToNot(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	cliPath = string(data)
	postgresRunner = helpers.Runner{
		Port: 5433 + GinkgoParallelNode(),
	}
	dbProcess = ifrit.Invoke(postgresRunner)
	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
	dbProcess.Signal(os.Interrupt)
	Eventually(dbProcess.Wait(), 10*time.Second).Should(Receive())
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	postgresRunner.CreateTestDB()
})

var _ = AfterEach(func() {
	postgresRunner.DropTestDB()
})
