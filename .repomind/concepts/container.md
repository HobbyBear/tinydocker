---
name: "容器"
description: "TinyDocker 中的核心业务对象——一个运行在独立命名空间中、受 cgroups 资源限制、拥有独立网络和文件系统的进程。与"虚拟机"容易混淆：容器共享宿主机内核，没有自己的 OS。"
---

# 概念：容器

## 是什么

TinyDocker 中的"容器"是一个运行在 Linux 命名空间隔离环境中的进程，拥有独立的 UTS/PID/Mount/Network/IPC 命名空间、受 cgroups 限制的 CPU 和内存配额、通过 veth+bridge 接入的网络、以及基于 OverlayFS 的独立根文件系统。

## 为什么有

TinyDocker 是一个教学项目，目的是通过约 500 行 Go 代码展示容器化的核心原理。容器的定义直接体现了：namespace（隔离）、cgroups（限制）、rootfs（文件系统）三个容器化基石。

## 用户侧表现

用户通过 `./tinydocker run <容器名> <可执行文件>` 启动一个容器。容器内的进程看到的是独立的文件系统（基于 Ubuntu 16.04 base image）、独立的网络栈（通过 bridge 出网）和受限的 CPU/内存资源（0.5 核 / 200MB）。容器退出后相关资源自动清理。

## 系统侧数据流

1. 用户执行 `./tinydocker run <name> <cmd>` → `main()` run 分支
2. 父进程 fork 子进程（设置 Cloneflags 指定命名空间隔离范围）
3. 父进程 → cgroups 模块写入 `/sys/fs/cgroup/{cpu,memory}/tinydocker/<name>/` 限制资源
4. 父进程 → network 模块分配 IP、创建 veth、配置子进程网络命名空间
5. 子进程 → workspace 模块通过 overlay mount + pivot_root 切换根文件系统
6. 子进程 → syscall.Exec 执行用户命令
7. 容器退出 → 父进程清理 cgroups 路径和 mount namespace

## 核心规则

- 容器名唯一：同一时间只能运行一个同名容器（cgroup 路径和 mount 路径都基于容器名）
- 网络类型固定：当前只支持 bridge 网络
- 镜像固定：使用 `ubuntu-base-16.04.6-base-amd64` 作为只读底层
- 资源限制硬编码：CPU 0.5 核，内存 200MB，不可配置

## 易混淆概念

- **不是虚拟机**：容器与宿主机共享内核，没有自己的 OS、没有模拟硬件
- **不是 Docker 容器**：TinyDocker 是简化实现，没有镜像仓库、Dockerfile、容器编排等概念
- **不是 chroot**：容器使用了 pivot_root + namespace 实现更彻底的隔离，chroot 只改根目录不改命名空间
