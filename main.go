package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"tinydocker/workspace"
)

// ./tinydocker run 容器名  可执行文件名
func main() {

	switch os.Args[1] {
	case "run":
		// 在一个新的命名空间
		initCmd, err := os.Readlink("/proc/self/exe")
		if err != nil {
			fmt.Println("get init process error ", err)
			return
		}
		containerName := os.Args[2]
		os.Args[1] = "init"
		cmd := exec.Command(initCmd, os.Args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
				syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		}
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
		workspace.DelMntNamespace(containerName)
		return
	case "init":
		var (
			containerName = os.Args[2]
			cmd           = os.Args[3]
		)
		if err := workspace.SetMntNamespace(containerName); err != nil {
			fmt.Println(err)
			return
		}
		syscall.Chdir("/")
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
		err := syscall.Exec(cmd, os.Args[3:], os.Environ())
		if err != nil {
			fmt.Println("exec proc fail ", err)
			return
		}
		fmt.Println("forever exec it ")
		return
	default:
		fmt.Println("not valid cmd")
	}
}
