package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	mntPath        = "/root/mnt"
	workLayerPath  = "/root/work"
	writeLayerPath = "/root/wlayer"
	imagePath      = "ubuntu-base-16.04.6-base-amd64"
	mntOldPath     = ".old"
)

func workerLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", workLayerPath, containerName)
}

func mntLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", mntPath, containerName)
}

func writeLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", writeLayerPath, containerName)
}

func mntOldLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", mntLayer(containerName), mntOldPath)
}

func SetMntNamespace(containerName string) error {
	if err := os.MkdirAll(mntLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir mntlayer fail err=%s", err)
	}
	if err := os.MkdirAll(workerLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir work layer fail err=%s", err)
	}
	if err := os.MkdirAll(writeLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir write layer fail err=%s", err)
	}

	if err := syscall.Mount("overlay", mntLayer(containerName), "overlay", 0,
		fmt.Sprintf("upperdir=%s,lowerdir=%s,workdir=%s",
			writeLayer(containerName), imagePath, workerLayer(containerName))); err != nil {
		return fmt.Errorf("mount overlay fail err=%s", err)
	}

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("reclare rootfs private fail err=%s", err)
	}

	if err := syscall.Mount(mntLayer(containerName), mntLayer(containerName), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs in new mnt space fail err=%s", err)
	}
	if err := os.MkdirAll(mntOldLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir mnt old layer fail err=%s", err)
	}
	if err := syscall.PivotRoot(mntLayer(containerName), mntOldLayer(containerName)); err != nil {
		return fmt.Errorf("pivot root  fail err=%s", err)
	}
	return nil
}

func delMntNamespace(path string) error {
	_, err := exec.Command("umount", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount fail path=%s err=%s", path, err)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove dir fail path=%s err=%s", path, err)
	}
	return nil
}

func DelMntNamespace(containerName string) error {
	if err := delMntNamespace(mntLayer(containerName)); err != nil {
		return err
	}
	if err := delMntNamespace(workerLayer(containerName)); err != nil {
		return err
	}
	if err := delMntNamespace(writeLayer(containerName)); err != nil {
		return err
	}
	return nil
}
