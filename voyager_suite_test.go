package voyager_test

import (
	"os"
	"time"

	"github.com/concourse/voyager/helpers"
	"github.com/gobuffalo/packr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"

	"testing"
)

func TestVoyager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Voyager Suite")
}

var postgresRunner helpers.Runner
var dbProcess ifrit.Process

var _ = BeforeSuite(func() {
	postgresRunner = helpers.Runner{
		Port: 5433 + GinkgoParallelNode(),
	}
	dbProcess = ifrit.Invoke(postgresRunner)
})

var _ = BeforeEach(func() {
	postgresRunner.CreateTestDB()
})

var _ = AfterEach(func() {
	postgresRunner.DropTestDB()
})

var _ = AfterSuite(func() {
	dbProcess.Signal(os.Interrupt)
	Eventually(dbProcess.Wait(), 10*time.Second).Should(Receive())
})

var box = packr.NewBox("./migrations")
var asset = box.MustBytes
var assetNames = func() []string {
	migrations := []string{}
	for _, name := range box.List() {
		if name != "migrations.go" {
			migrations = append(migrations, name)
		}
	}

	return migrations
}
