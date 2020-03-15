package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"path/filepath"
	"strconv"
	"io/ioutil"
)

// docker run <container> cmd args
// go run main.go run cmd args
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run () {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr {
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
                UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
	        },
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())

	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))

	must(cmd.Run())
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
        pids := filepath.Join(cgroups, "pids")

	must(os.Mkdir(filepath.Join(pids, "liz"), 0755))
	must(ioutil.WriteFile(filepath.Join(pids, "liz/pids.max"), []byte("20"),0700))
	//Removes the news cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "liz/notify_on_release"), []byte("1"),0700))
	must(ioutil.WriteFile(filepath.Join(pids, "liz/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}



func must(err error) {
	if err != nil {
		panic(err)
	}
}
