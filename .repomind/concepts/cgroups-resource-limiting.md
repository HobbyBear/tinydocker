---
name: "Cgroups 资源限制"
description: "通过 Linux cgroups v1 对容器的 CPU 和内存进行硬限制的机制。涉及 cgroup 文件系统路径、CPU 配额（cfs_quota_us）和内存限制（limit_in_bytes）的配置与清理。"
---

# 概念：Cgroups 资源限制

## 是什么

通过 Linux cgroups v1 的 cpu 和 memory 子系统，限制容器进程可使用的 CPU 时间和物理内存上限。CPU 限制为 50000μs/周期（约 0.5 核），内存限制为 200MB。这是防止单个容器耗尽宿主机资源的保护机制。

## 为什么有

容器共享宿主机内核和硬件资源，没有 cgroups 限制的容器可以无限制使用 CPU 和内存，影响宿主机和其他容器的稳定性。TinyDocker 通过 cgroups 展示了容器资源隔离的另一半：不只隔离"看见什么"，还要限制"用多少"。

## 用户侧表现

用户在容器内运行的程序受 CPU 和内存上限约束。如果程序尝试使用超过 200MB 内存，会被 OOM killer 终止。如果 CPU 密集计算，最多使用约 0.5 个核心的计算能力。用户看不到 cgroup 文件系统本身（在宿主机上操作）。

## 系统侧数据流

1. 父进程在 `cmd.Start()` 后拿到子进程 PID
2. `cgroups.ConfigDefaultCgroups(pid, containerName)` 被调用
3. 在 `/sys/fs/cgroup/cpu/tinydocker/<name>/` 下创建目录
4. 写入 `cpu.cfs_quota_us` = "50000"（限制 CPU 使用）
5. 写入 `tasks` = pid（将进程纳入 cgroup 控制）
6. 在 `/sys/fs/cgroup/memory/tinydocker/<name>/` 下重复类似操作
7. 写入 `memory.limit_in_bytes` = "200m"
8. 容器退出后 `CleanCgroupsPath()` 通过 `cgdelete` 命令清理 cgroup 层级

## 核心规则

- CPU 限制类型：CFS 带宽控制（cpu.cfs_quota_us），非 CPU 集或份额
- 默认周期 100ms，限额 50ms → 0.5 核
- 内存限制值 "200m" 中的 m 表示 MiB（1048576 bytes）
- cgroup 路径结构：`/sys/fs/cgroup/{subsystem}/tinydocker/{containerName}`
- 清理依赖外部命令 `cgdelete`（来自 libcgroup-tools 包）

## 易混淆概念

- **不是 namespace**：namespace 决定"看见什么"，cgroups 决定"用多少"
- **不是 Docker 的 --cpus/--memory**：TinyDocker 硬编码限制值，Docker 可配置
- **不是 cgroup v2**：TinyDocker 使用 cgroup v1 的 per-subsystem 层级结构
