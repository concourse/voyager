package integration_test

import (
	"github.com/onsi/gomega/gexec"
	"os"
	"time"

	"github.com/ddadlani/voyager/helpers"
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
	postgresRunner = helpers.Runner{
		Port: 5433 + GinkgoParallelNode(),
	}

	dbProcess = ifrit.Invoke(postgresRunner)

	binPath, err := gexec.Build("github.com/ddadlani/voyager/cli")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	cliPath = string(data)

	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
	dbProcess.Signal(os.Interrupt)
	Eventually(dbProcess.Wait(), 10*time.Second).Should(Receive())
})


var _ = BeforeEach(func() {
	postgresRunner.CreateTestDB()
})

var _ = AfterEach(func() {
	postgresRunner.DropTestDB()
})

