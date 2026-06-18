---
name: "Cgroups 资源限制模块"
description: "容器 CPU 和内存的硬件资源限制模块。用于定位容器资源配额的配置入口、默认限制值、以及容器销毁时的 cgroup 清理逻辑。与容器生命周期模块紧密联动。"
keywords:
- "cgroups"
- "资源限制"
- "CPU"
- "内存"
- "memory"
- "resource limit"
- "cgroup"
- "Cgroups 资源限制模块"
---

# Cgroups 资源限制模块

## 业务描述

通过 Linux cgroups 机制对容器进程施加 CPU 和内存资源限制。当前实现为固定配额：CPU 限制为 50000μs（即 50ms/100ms 周期，约 0.5 核），内存限制为 200MB。容器销毁时通过 `cgdelete` 命令清理 cgroup 层级。

## 关键代码

- `cgroups/cgroup.go:17 ConfigDefaultCgroups()` — 创建 cgroup 目录并写入资源限制
- `cgroups/cgroup.go:49 CleanCgroupsPath()` — 容器退出后清理 cgroup 路径

## 常见修改场景

- 调整默认 CPU/内存限制：修改 `cpu.cfs_quota_us` 和 `memory.limit_in_bytes` 的写入值
- 新增资源类型限制（如 blkio）：参考 `ConfigDefaultCgroups` 的模式新增子目录和写入
- 改为可配置限制：需要从 main.go 传入参数，而非硬编码

## AI 注意事项

- cgroup 路径结构为 `/sys/fs/cgroup/{subsystem}/tinydocker/{containerName}`，依赖 cgroup v1 文件系统
- CPU 限制使用的是 `cpu.cfs_quota_us`（CFS 带宽控制），需要确认内核支持
- 清理使用外部命令 `cgdelete` 而非直接删除目录，如果系统未安装 libcgroup-tools 会导致清理失败（当前只 log error 不返回错误）
- 与容器生命周期模块是紧耦合：`ConfigDefaultCgroups` 在容器启动后调用，`CleanCgroupsPath` 在容器退出后调用
