---
name: "命名空间隔离"
description: "Linux namespace 机制在 TinyDocker 中的运用。涉及 UTS/PID/Mount/Network/IPC 五种命名空间的创建、切换和数据流动。与"cgroups"容易混淆：namespace 管"看见什么"，cgroups 管"用多少"。"
---

# 概念：命名空间隔离

## 是什么

Linux 内核的 namespace 机制允许多个进程组拥有独立的系统资源视图。TinyDocker 使用 5 种命名空间创建隔离环境：UTS（主机名）、PID（进程树）、Mount（文件系统挂载点）、Network（网络栈）、IPC（进程间通信）。

## 为什么有

命名空间是容器隔离的核心技术。TinyDocker 的教学目标之一就是展示如何用 Linux 原生能力（而非虚拟机）实现进程隔离。5 种命名空间的组合覆盖了容器最基本的隔离需求。

## 用户侧表现

用户无直接感知。容器内的进程看到的 PID 从 1 开始、主机名独立、拥有自己的网卡和路由表、文件系统根目录独立——这些都是命名空间隔离的效果。

## 系统侧数据流

1. `main.go` run 分支通过 `syscall.SysProcAttr.Cloneflags` 指定创建命名空间：
   - `CLONE_NEWUTS` — 隔离主机名和域名
   - `CLONE_NEWPID` — 隔离进程 ID 空间
   - `CLONE_NEWNS` — 隔离挂载点
   - `CLONE_NEWNET` — 隔离网络栈
   - `CLONE_NEWIPC` — 隔离 SysV IPC 和 POSIX 消息队列
2. Mount namespace 进一步由 `workspace.SetMntNamespace()` 通过 overlay mount + pivot_root 定制
3. Network namespace 由 `network.setContainerIp()` 通过 `enterContainerNetns()` 切换后配置

## 核心规则

- 所有 5 种命名空间在 fork 时一次性创建，不可动态增减
- Network namespace 创建时子进程无网络，需父进程外部配置
- Mount namespace 创建后子进程先 pivot_root 再执行用户命令，确保隔离生效
- 父进程不进入子进程的命名空间（网络配置除外，通过 netns 切换实现）

## 易混淆概念

- **不是 cgroups**：命名空间控制"能看到什么"（隔离），cgroups 控制"能用多少"（限制）
- **不是 chroot**：chroot 只改文件系统根，不改 PID/网络/挂载点名空间
- **不是 Docker/K8s 的 namespace**：Docker/K8s 中的 "namespace" 也可能指逻辑分组或租户隔离，与 Linux namespace 不同
