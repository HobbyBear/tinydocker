---
name: "IPAM（IP 地址管理）"
description: "容器 IP 地址的分配、回收和持久化管理。基于 bitmap 位图跟踪子网中每个 IP 的使用状态，通过 JSON 文件持久化到磁盘。用于定位 IP 分配冲突、子网耗尽和状态恢复问题。"
---

# 概念：IPAM（IP 地址管理）

## 是什么

IP Address Management 负责从给定的 CIDR 子网中为容器分配独立 IP 地址。核心数据结构是 bitmap（位图），每一位代表子网中的一个 IP 地址。已分配的 IP 对应位设为 1，释放时清 0。分配状态通过 JSON 文件持久化到 `/root/subnet.json`，重启后可以恢复。

## 为什么有

每个容器需要独立 IP 才能在同一 bridge 网络中通信。如果每次都随机分配或手动指定，容易出现 IP 冲突。IPAM 提供了一种简单可靠的 IP 管理方案：自动分配、避免冲突、持久化存储、支持回收复用。

## 用户侧表现

用户无直接感知。每次创建容器时自动获得一个未被占用的 IP。如果子网中所有可用 IP 都已分配完（约 253 个可用地址，排除网络号和广播地址），新容器创建会失败。

## 系统侧数据流

1. `Init()` → `IpAmfs.SetIpUsed(defaultSubnet)` 标记网关 IP（.1）已占用
2. 创建容器 → `IpAmfs.AllocIp(defaultSubnet)` 扫描 bitmap 找第一个未用位
3. bitmap 位从 1 开始扫描（pos=0 是网络号，不分配），到 (2^(total-ones) - 2) 结束（广播地址不分配）
4. 找到未用位 → `bitmap.BitSet(pos)` 标记 → `ipamfs.sync()` 写回 JSON
5. 释放 IP → `IpAmfs.ReleaseIp(subnet, ip)` → `bitmap.BitClean(pos)` → `sync()`
6. 重启恢复 → `ipamfs.loadConf()` 从 JSON 反序列化 bitmap 状态

## 核心规则

- 实现方式：文件系统存储的 JSON + 内存 bitmap
- 分配策略：顺序扫描，取第一个未用位（非随机、非最优适配）
- 网络号（pos=0）和广播地址（pos=total-1）永远不可分配
- bitmap 大小 = 2^(total - ones)，即子网中的总 IP 数
- IP ↔ uint32 转换：BigEndian，仅 IPv4
- 持久化路径：`/root/subnet.json`（通过 `config.IpAmStorageFsPath`）
- 加载时 JSON 为空或文件不存在 → 初始化为空 bitmap，不报错

## 易混淆概念

- **不是 DHCP**：DHCP 是动态主机配置协议，有租约、续约、广播发现等机制；TinyDocker IPAM 是本地 bitmap 管理
- **不是 Docker IPAM**：Docker 支持多种 IPAM driver（default、null 等），TinyDocker 只有文件系统实现
- **不是子网划分**：IPAM 不创建新子网，只在已有子网中分配/回收 IP
