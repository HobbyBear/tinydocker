package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ./tinydocker run  可执行文件名
func main() {

	switch os.Args[1] {
	case "run":
		// 在一个新的命名空间
		fmt.Println("run pid ", os.Getpid(), "ppid", os.Getppid())
		initCmd, err := os.Readlink("/proc/self/exe")
		if err != nil {
			fmt.Println("get init process error ", err)
			return
		}
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
		fmt.Println("init proc end", initCmd)
		return
	case "init":
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("pwd", err)
			return
		}
		path := pwd + "/ubuntu-base-16.04.6-base-amd64"
		fmt.Println(path)
		// systemd 为init进程时，挂载默认是共享模式挂载的，共享模式挂载会让所有命名空间都能看到各自的挂载的目录
		// 后续调用pivot root会失败，所以将命名空间声明为私有的，MS_REC是mount选项中的一个标志，用于递归地挂载一个目录及其所有子目录
		syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
		//syscall.Chroot("./ubuntu-base-16.04.6-base-amd64")

		if err := syscall.Mount(path, path, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			fmt.Println("Mount", err)
			return
		}
		if err := os.MkdirAll(path+"/.old", 0700); err != nil {
			fmt.Println("mkdir", err)
			return
		}
		fmt.Println(path, path+"/.old")
		err = syscall.PivotRoot(path, path+"/.old")
		if err != nil {
			fmt.Println("pivot root ", err)
			return
		}
		syscall.Chdir("/")

		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		//syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
		syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

		cmd := os.Args[2]
		fmt.Println("will exec cmd=", cmd)
		err = syscall.Exec(cmd, os.Args[2:], os.Environ())
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
