---
name: "OverlayFS 联合文件系统"
description: "基于 Linux OverlayFS 的容器 rootfs 实现。涉及 lowerdir（只读镜像）、upperdir（可写层）、workdir 三层模型和 pivot_root 根切换。与"volume/bind mount"容易混淆：overlay 提供写时复制，bind mount 只是路径映射。"
---

# 概念：OverlayFS 联合文件系统

## 是什么

利用 Linux OverlayFS 将宿主机的只读 Ubuntu 基础镜像（lowerdir）与容器的可写层（upperdir）联合挂载到同一个目录，使容器"看到"一个完整的可写文件系统，而对文件的修改实际上只写入可写层，不会污染基础镜像。

## 为什么有

容器镜像通常是只读的，但容器内的进程需要写入文件（如日志、临时文件）。OverlayFS 的写时复制（copy-on-write）机制允许多个容器共享同一个只读镜像，各自的写入独立存储，实现存储效率和隔离的平衡。TinyDocker 用约 20 行代码展示这一核心原理。

## 用户侧表现

用户在容器内看到的文件系统是完整的 Ubuntu 16.04 目录树（/bin, /etc, /usr 等），可以自由创建、修改、删除文件。容器退出后所有修改丢失（除非 mount 目录被外部保留）。

## 系统侧数据流

1. 子进程调用 `workspace.SetMntNamespace(containerName)`
2. 创建三个目录：
   - mntLayer: `/root/mnt/<name>` — 联合挂载点（容器看到的内容）
   - workerLayer: `/root/work/<name>` — overlay 工作目录
   - writeLayer: `/root/wlayer/<name>` — 可写层（upperdir）
3. 执行 `mount("overlay", mntLayer, "overlay", 0, "upperdir=writeLayer,lowerdir=imagePath,workdir=workerLayer")`
4. `imagePath` = `ubuntu-base-16.04.6-base-amd64`（只读底层）
5. 递归将根设为私有挂载：`mount("", "/", "", MS_PRIVATE|MS_REC, "")`
6. bind mount 将 mntLayer 绑定到自身
7. 执行 `pivot_root(mntLayer, mntOldLayer)` 切换容器的根文件系统

## 核心规则

- lowerdir 只读，upperdir 可读写，workdir 必须与 upperdir 在同一文件系统
- pivot_root 之前必须 MS_PRIVATE 递归设置，防止挂载事件传播到宿主机
- 当前镜像路径为相对路径，依赖运行时的 workdir
- 容器退出后 `DelMntNamespace()` 依次 umount 并删除三层目录

## 易混淆概念

- **不是 volume/bind mount**：bind mount 只是把宿主机目录映射进容器，无写时复制
- **不是 Device Mapper/AUFS**：OverlayFS 是另一种联合文件系统实现，自 Linux 3.18 进入主线
- **不是 Docker overlay2 驱动**：TinyDocker 直接调用 syscall.Mount，没有 Docker 的多层镜像缓存和 digest 管理
