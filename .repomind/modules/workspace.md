---
name: "容器文件系统模块"
description: "容器 rootfs 和 mount namespace 管理模块。负责 overlay 联合挂载、pivot_root 根切换和容器退出后的挂载点清理。用于定位容器文件系统隔离的配置和问题排查。"
keywords:
- "workspace"
- "rootfs"
- "overlay"
- "mount"
- "pivot_root"
- "文件系统"
- "filesystem"
- "mnt"
- "容器文件系统模块"
---

# 容器文件系统模块

## 业务描述

为容器提供独立的文件系统视图。基于 OverlayFS 实现写时复制：将宿主机上的只读镜像（lowerdir）与容器的可写层（upperdir）联合挂载，通过 pivot_root 将容器进程的根目录切换到该挂载点。容器退出后卸载所有挂载点并清理目录。

## 关键代码

- `workspace/workspace.go:34 SetMntNamespace()` — 设置容器文件系统：创建目录 → overlay mount → 递归设私 → bind mount → pivot_root
- `workspace/workspace.go:78 DelMntNamespace()` — 容器退出后清理：卸载 → 删除三层目录
- `workspace/workspace.go:18-32` — 各层路径生成函数（workerLayer, mntLayer, writeLayer, mntOldLayer）

## 常见修改场景

- 更换基础镜像：修改 `imagePath` 常量指向新的 rootfs 目录
- 调整存储路径：修改 `/root/mnt`、`/root/work`、`/root/wlayer` 路径常量
- 排查"找不到文件"问题：先确认 imagePath 指向的 rootfs 目录存在且完整
- 新增 volume 挂载：在 pivot_root 之后添加 bind mount 逻辑

## AI 注意事项

- OverlayFS 需要 lowerdir 目录可读、upperdir 和 workdir 在同一文件系统，修改路径时需遵守这些约束
- pivot_root 之前必须先 `MS_PRIVATE|MS_REC` 递归将根文件系统设为私有，否则会影响到宿主机
- 当前 imagePath 为相对路径 `ubuntu-base-16.04.6-base-amd64`，运行时依赖工作目录正确
- `DelMntNamespace` 清理时使用外部 `umount` 命令而非 `syscall.Unmount`，注意依赖关系
- 容器退出后清理失败会留下挂载点和目录残渣，但没有重试或告警机制
