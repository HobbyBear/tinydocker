---
name: "Bridge 容器网络"
description: "基于 Linux bridge + veth pair 的容器网络方案。涉及网桥创建、veth 对连接、容器 IP 配置、SNAT 出网和网络命名空间切换。与"host 网络""overlay 网络"容易混淆：bridge 是独立的二层虚拟网络。"
---

# 概念：Bridge 容器网络

## 是什么

TinyDocker 通过在宿主机上创建 Linux bridge 设备，再为每个容器创建一对 veth（Virtual Ethernet）设备（一端插到 bridge，另一端放入容器的网络命名空间），实现容器与宿主机的网络互通。配合 iptables SNAT 规则，容器可以访问外部网络。

## 为什么有

容器创建时拥有独立的 Network namespace（通过 CLONE_NEWNET），此时容器内只有一个 loopback 接口，无法访问外部网络。Bridge + veth 是 Linux 上最经典的容器网络方案，TinyDocker 用它演示容器网络的完整链路。

## 用户侧表现

容器内的进程看到的是"一台有网卡的机器"：拥有独立 IP 地址（从默认子网 `192.169.0.1/24` 中分配）、默认路由指向 bridge 网关、可以通过 SNAT 访问外网。用户可在容器内直接 curl/ping 外部地址。

## 系统侧数据流

1. 网络初始化：`Init()` → `BridgeDriver.CreateNetwork("testbridge", "192.169.0.1/24", "bridge")`
2. 创建 Linux bridge：`createBridge()` → `netlink.LinkAdd(br)` + `AddrAdd` + `LinkSetUp`
3. 设置 SNAT：`setSNat()` → iptables `-t nat -A POSTROUTING -s <subnet> ! -o <bridge> -j MASQUERADE`
4. 标记网关 IP 已用：`IpAmfs.SetIpUsed(defaultSubnet)`
5. 为容器分配 IP：`IpAmfs.AllocIp(defaultSubnet)` → bitmap 查找未用 IP
6. 创建 veth pair：`BridgeDriver.CrateVeth()` → `netlink.LinkAdd(vethLink)`，一端连 bridge
7. 进入容器网络命名空间：`enterContainerNetns()` → `netns.Set()` 切换当前线程的 netns
8. 在容器内配置 veth：`setContainerIp()` → 设 IP、设路由（默认网关）、up loopback
9. 恢复宿主网络命名空间，发送 SIGUSR2 通知子进程网络就绪

## 核心规则

- 仅支持 bridge 网络类型（networktype = "bridge"）
- 默认子网 `192.169.0.1/24`（网关 IP 即 .1，不可分配给容器）
- veth 命名规则：`veth-{随机数}-{网络名}` / `cif-veth-{随机数}-{网络名}`
- bridge 命名规则：`br-{网络名}`，截断到 15 字符
- 进入容器 netns 必须 LockOSThread，离开时必须 Unlock + Close
- 网络配置持久化到 `/root/network.json`（跨重启恢复）

## 易混淆概念

- **不是 host 网络**：host 网络共享宿主机网络栈，不隔离；bridge 有独立 netns
- **不是 overlay/VXLAN**：overlay 网络跨主机，TinyDocker 的 bridge 只在本机
- **不是 Docker bridge 驱动**：TinyDocker 直接操作 netlink，没有 Docker 的 iptables 端口映射和 DNS 解析
