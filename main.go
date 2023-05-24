package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"tinydocker/config"
	"tinydocker/log"
	"tinydocker/network"
	"tinydocker/workspace"
)

// ./tinydocker run 容器名  可执行文件名
func main() {
	switch os.Args[1] {
	case "run":
		if err := network.Init(); err != nil {
			log.Error("net work fail err=%s", err)
			return
		}
		fmt.Println(config.Title())
		// 在一个新的命名空间
		initCmd, err := os.Readlink("/proc/self/exe")
		if err != nil {
			log.Error("get init process error %s", err)
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
		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		// 等待子进程完全启动
		time.Sleep(2 * time.Second)
		if err := network.ConfigDefaultNetworkInNewNet(cmd.Process.Pid); err != nil {
			log.Error("config network fail %s", err)
		}
		cmd.Wait()
		workspace.DelMntNamespace(os.Args[2])
		return
	case "init":
		var (
			containerName = os.Args[2]
			cmd           = os.Args[3]
		)
		log.Info("Wait  SIGUSR2 signal arrived ....")
		// 等待父进程网络命名空间设置完毕
		network.WaitParentSetNewNet()
		if err := workspace.SetMntNamespace(containerName); err != nil {
			log.Error("SetMntNamespace %s", err)
			return
		}
		syscall.Chdir("/")
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
		err := syscall.Exec(cmd, os.Args[3:], os.Environ())
		if err != nil {
			log.Error("exec proc fail %s", err)
			return
		}
		log.Error("forever not  exec it ")
		return
	default:
		log.Error("not valid cmd")
	}
}
