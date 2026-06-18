---
name: "容器网络模块"
description: "容器网络模块，负责 Linux bridge 创建、veth pair 连接、IP 地址分配（IPAM）和 SNAT 出网。用于定位容器网络初始化、IP 分配流程和网络命名空间切换逻辑。"
keywords:
- "网络"
- "network"
- "bridge"
- "veth"
- "IPAM"
- "SNAT"
- "bridge network"
- "network namespace"
- "容器网络模块"
---

# 容器网络模块

## 业务描述

为容器提供网络连接能力。核心流程：在宿主机上创建 Linux bridge → 通过 IPAM 为容器分配 IP → 创建 veth pair 连接容器命名空间和 bridge → 配置 SNAT 实现容器出网。网络配置通过 JSON 文件持久化到 `/root/network.json`，IP 分配状态持久化到 `/root/subnet.json`。

## 关键代码

- `network/network.go:89 Init()` — 网络模块初始化，创建默认 bridge 网络并标记网关 IP 已用
- `network/network.go:100 ConfigDefaultNetworkInNewNet()` — 为新容器分配 IP、创建 veth、配置容器侧网络
- `network/bridge_network.go:72 CreateNetwork()` — 创建 bridge 网络（含 bridge 设备、SNAT 规则、持久化）
- `network/bridge_network.go:123 CrateVeth()` — 创建 veth pair 并挂到 bridge 上
- `network/bridge_network.go:155 setContainerIp()` — 进入容器网络命名空间设置 IP 和默认路由
- `network/ipam_fs.go:43 AllocIp()` — 从子网中分配一个未使用 IP（基于 bitmap）
- `network/ipam_fs.go:22 SetIpUsed()` — 标记 IP 已被占用
- `network/ipam_fs.go:76 ReleaseIp()` — 释放 IP
- `network/network.go:30 Sync() / LoadConf()` — 网络配置的持久化与加载

## 常见修改场景

- 新增网络驱动类型（非 bridge）：需要在 `networktype` 常量中新增类型并实现对应 driver 接口
- 调整默认子网：修改 `defaultSubnet` 常量（当前为 `192.169.0.1/24`）
- 修改 IP 分配策略：IPAM 当前是顺序扫描 bitmap，改策略先看 `AllocIp()`
- 排查网络不通：先看 veth 是否正确创建并挂到 bridge，再看容器内的 IP 和路由是否设置正确

## AI 注意事项

- 网络配置和 IPAM 数据分别持久化到两个 JSON 文件，修改结构体时需同步处理序列化兼容性
- `enterContainerNetns()` 使用 `runtime.LockOSThread()` 锁定 OS 线程后切换网络命名空间，这是 Linux namespace 操作的关键约束
- `setContainerIp()` 内部通过 defer 机制自动恢复原始命名空间，修改此函数时不要破坏 defer 链
- 父子进程网络同步依赖 `SIGUSR2` 信号：父进程配置完网络后发信号给子进程，子进程收到信号后继续执行
- 默认子网 `192.169.0.1/24` 不是标准的私有地址范围（标准是 192.168.x.x），可能是笔误
