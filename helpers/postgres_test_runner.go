package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type Runner struct {
	Port int
}

func (runner Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	pgBase := filepath.Join(os.TempDir(), "test-pg-runner")

	err := os.MkdirAll(pgBase, 0755)
	Expect(err).ToNot(HaveOccurred())

	tmpdir, err := ioutil.TempDir(pgBase, "postgres")
	Expect(err).ToNot(HaveOccurred())

	currentUser, err := user.Current()
	Expect(err).ToNot(HaveOccurred())

	initdbPath, err := exec.LookPath("initdb")
	Expect(err).ToNot(HaveOccurred())

	postgresPath, err := exec.LookPath("postgres")
	Expect(err).ToNot(HaveOccurred())

	initCmd := exec.Command(initdbPath, "-U", "postgres", "-D", tmpdir, "-E", "UTF8", "--no-local")
	startCmd := exec.Command(postgresPath, "-k", "/tmp", "-D", tmpdir, "-h", "127.0.0.1", "-p", strconv.Itoa(runner.Port))

	if currentUser.Uid == "0" {
		pgUser, err := user.Lookup("postgres")
		Expect(err).ToNot(HaveOccurred())

		var uid, gid uint32
		_, err = fmt.Sscanf(pgUser.Uid, "%d", &uid)
		Expect(err).ToNot(HaveOccurred())

		_, err = fmt.Sscanf(pgUser.Gid, "%d", &gid)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chown(tmpdir, int(uid), int(gid))
		Expect(err).ToNot(HaveOccurred())

		credential := &syscall.Credential{Uid: uid, Gid: gid}

		initCmd.SysProcAttr = &syscall.SysProcAttr{}
		initCmd.SysProcAttr.Credential = credential

		startCmd.SysProcAttr = &syscall.SysProcAttr{}
		startCmd.SysProcAttr.Credential = credential
	}

	session, err := gexec.Start(
		initCmd,
		gexec.NewPrefixedWriter("[o][initdb] ", GinkgoWriter),
		gexec.NewPrefixedWriter("[e][initdb] ", GinkgoWriter),
	)
	Expect(err).ToNot(HaveOccurred())

	<-session.Exited

	Expect(session).To(gexec.Exit(0))

	ginkgoRunner := &ginkgomon.Runner{
		Name:          "postgres",
		Command:       startCmd,
		AnsiColorCode: "90m",
		StartCheck:    "database system is ready to accept connections",
		Cleanup: func() {
			os.RemoveAll(tmpdir)
		},
	}

	return ginkgoRunner.Run(signals, ready)
}

func (runner *Runner) DataSourceName() string {
	return fmt.Sprintf("host=/tmp user=postgres dbname=testdb sslmode=disable port=%d", runner.Port)
}

func (runner *Runner) CreateTestDB() {
	createdb := exec.Command("createdb", "-h", "/tmp", "-U", "postgres", "-p", strconv.Itoa(runner.Port), "testdb")

	createS, err := gexec.Start(createdb, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	<-createS.Exited

	if createS.ExitCode() != 0 {
		runner.DropTestDB()

		createdb := exec.Command("createdb", "-h", "/tmp", "-U", "postgres", "-p", strconv.Itoa(runner.Port), "testdb")
		createS, err = gexec.Start(createdb, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	}

	<-createS.Exited

	Expect(createS).To(gexec.Exit(0))
}

func (runner *Runner) DropTestDB() {
	dropdb := exec.Command("dropdb", "-h", "/tmp", "-U", "postgres", "-p", strconv.Itoa(runner.Port), "testdb")
	dropS, err := gexec.Start(dropdb, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	<-dropS.Exited

	Expect(dropS).To(gexec.Exit(0))
}
