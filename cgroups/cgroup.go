package cgroups

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"tinydocker/log"
)

const (
	cgroupsPath = "/sys/fs/cgroup"
	dockerName  = "tinydocker"
)

func ConfigDefaultCgroups(pid int, containerName string) error {

	var (
		cpuPath    = path.Join(cgroupsPath, "cpu", dockerName, containerName)
		memoryPath = path.Join(cgroupsPath, "memory", dockerName, containerName)
	)

	// 创建容器的控制目录
	if err := os.MkdirAll(cpuPath, 0700); err != nil {
		return fmt.Errorf("create cgroup path fail err=%s", err)
	}
	if err := os.MkdirAll(memoryPath, 0700); err != nil {
		return fmt.Errorf("create cgroup path fail err=%s", err)
	}
	// 设置cpu
	if err := os.WriteFile(path.Join(cpuPath, "cpu.cfs_quota_us"), []byte("50000"), 0700); err != nil {
		return fmt.Errorf("write cpu quota us fail err=%s", err)
	}
	if err := os.WriteFile(path.Join(cpuPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("write cpu tasks  fail err=%s", err)
	}

	// 设置内存
	if err := os.WriteFile(path.Join(memoryPath, "memory.limit_in_bytes"), []byte("200m"), 0700); err != nil {
		return fmt.Errorf("write memory limit bytes fail err=%s", err)
	}
	if err := os.WriteFile(path.Join(memoryPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("write memory tasks  fail err=%s", err)
	}
	return nil
}

func CleanCgroupsPath(containerName string) error {
	output, err := exec.Command("cgdelete", "-r", fmt.Sprintf("memory:%s/%s", dockerName, containerName)).Output()
	if err != nil {
		log.Error("cgdelete fail err=%s output=%s", err, string(output))
	}
	output, err = exec.Command("cgdelete", "-r", fmt.Sprintf("cpu:%s/%s", dockerName, containerName)).Output()
	if err != nil {
		log.Error("cgdelete fail err=%s output=%s", err, string(output))
	}
	return nil
}
