---
name: "容器生命周期模块"
description: "TinyDocker 容器启动与销毁的入口编排模块。用于定位 run/init 两条执行路径、容器创建主流程和子进程通信机制。影响面横跨 cgroups、network、workspace 三个模块。"
keywords:
- "容器生命周期"
- "container lifecycle"
- "run"
- "init"
- "main"
- "容器启动"
- "容器销毁"
- "容器生命周期模块"
- "container-lifecycle"
---

# 容器生命周期模块

## 业务描述

负责容器从启动到销毁的完整生命周期编排。入口为 `main.go`，通过 `run` 和 `init` 两条命令分支实现父子进程协作：父进程负责创建命名空间隔离的子进程、配置 cgroups 资源限制和网络，子进程负责设置 mount namespace、切换根文件系统并执行用户命令。

## 关键代码

- `main.go:17 main()` — 唯一入口，按 `os.Args[1]` 分发 run/init 两条路径
- `main.go:20-59 run 分支` — 父进程逻辑：网络初始化 → fork 子进程 → cgroups 限制 → 网络配置 → 等待子进程 → 清理
- `main.go:60-81 init 分支` — 子进程逻辑：等待父进程网络就绪信号 → 挂载文件系统 → pivot root → 执行用户进程

## 常见修改场景

- 新增容器生命周期阶段（如 pre-start hook）：先看 `main()` 的 run 分支流程
- 修改父子进程通信机制：当前使用 `SIGUSR2` 信号 + 2 秒 sleep，改这里先看 `ConfigDefaultNetworkInNewNet()` 和 `noticeSunProcessNetConfigFin()`
- 新增命令（如 stop/exec）：在 `main()` 的 switch 中新增 case

## AI 注意事项

- run 和 init 运行在**不同进程**中，run 在父进程、init 在 fork 出的子进程，修改时需注意进程边界
- 父子进程间通信依赖 `SIGUSR2` 信号 + 2 秒 sleep，没有实际的同步机制，这是已知的脆弱点
- Cloneflags 决定了子进程的命名空间隔离范围（UTS/PID/Mount/Network/IPC），修改时需确认对下游模块的影响
- 清理流程在 `cmd.Wait()` 之后执行，如果 cmd.Wait 阻塞或子进程异常退出，清理可能被跳过
