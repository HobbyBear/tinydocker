package cgroups

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

const (
	cgroupsPath = "/sys/fs/cgroup"
)

func ConfigDefaultCgroups(pid int, containerName string) error {

	var (
		cpuPath    = path.Join(cgroupsPath, "cpu", "tinydocker", containerName)
		memoryPath = path.Join(cgroupsPath, "memory", "tinydocker", containerName)
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
	return os.RemoveAll(path.Join(cgroupsPath, containerName))
}
