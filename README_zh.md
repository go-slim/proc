# proc — Go 进程工具库

![CI](https://github.com/go-slim/proc/actions/workflows/ci.yml/badge.svg)

[English](README.md) | [简体中文](README_zh.md)

小巧、专注的工具库，用于处理当前进程、信号和运行子进程。

## 功能特性

- **进程信息**：通过 `Pid()`、`Name()`、`WorkDir()`、`Path(...)`、`Pathf(...)`、`Context()` 获取进程元数据
- **信号处理**：使用 `On()`/`Once()` 注册监听器，通过 `Cancel()` 移除，通过 `Notify()` 触发
- **优雅关闭**：使用 `Shutdown(syscall.Signal)` 优雅关闭，支持配置强制终止延迟（测试友好的存根设计）
- **命令执行**：运行外部命令，支持超时、环境变量、工作目录和生命周期回调
- **日志控制**：通过 `Logger` 变量控制调试输出

模块路径：`go-slim.dev/proc`

## 安装

```bash
go get go-slim.dev/proc
```

Go 版本要求：`1.24`（参见 `proc/go.mod`）

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"
    "syscall"
    "time"

    proc "go-slim.dev/proc"
)

func main() {
    // 获取进程信息
    fmt.Println("pid:", proc.Pid(), "name:", proc.Name(), "wd:", proc.WorkDir())

    // 构建相对于工作目录的路径
    configPath := proc.Path("config", "app.yaml")
    fmt.Println("config path:", configPath)

    // 监听 SIGTERM 信号（仅一次）
    proc.Once(syscall.SIGTERM, func() {
        fmt.Println("terminating...")
    })

    // 运行带超时和错误处理的命令
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := proc.Exec(ctx, proc.ExecOptions{
        Command: "sh",
        Args:    []string{"-c", "echo ok"},
    })
    if err != nil {
        log.Fatalf("command failed: %v", err)
    }
}
```

## 信号处理

信号 API 允许你为操作系统信号注册自定义处理器：

- **`On(sig, fn) uint32`** - 注册一个监听器，每次收到信号时都会触发。返回监听器 ID。
- **`Once(sig, fn) uint32`** - 注册一次性监听器，执行后自动移除。返回监听器 ID。
- **`Cancel(id...)`** - 通过 ID 移除监听器。可以安全地传入无效或已移除的 ID。
- **`Notify(sig) bool`** - 手动触发指定信号的回调。如果没有找到监听器则返回 false。

**自动关闭**：包在初始化时会为常见信号（`SIGHUP`、`SIGINT`、`SIGQUIT`、`SIGTERM`）安装信号监听器，触发优雅关闭。

### 示例：自定义信号处理

```go
import (
    "fmt"
    "syscall"
    proc "go-slim.dev/proc"
)

// 注册可重复触发的处理器
id := proc.On(syscall.SIGUSR1, func() {
    fmt.Println("Received SIGUSR1")
})

// 注册一次性处理器
proc.Once(syscall.SIGUSR2, func() {
    fmt.Println("This will only run once")
})

// 稍后：取消重复触发的处理器
proc.Cancel(id)
```

## 优雅关闭

关闭系统提供了可配置强制终止行为的优雅终止功能。

```go
import (
    "syscall"
    "time"
    proc "go-slim.dev/proc"
)

// 可选：设置强制终止前的延迟
proc.SetTimeToForceQuit(2 * time.Second)

// 触发优雅关闭序列
err := proc.Shutdown(syscall.SIGTERM)
if err != nil {
    // 处理错误
}
```

**行为说明**：
- 如果调用 `SetTimeToForceQuit()` 设置的延迟 > 0：
  1. 在 goroutine 中调用 `Notify(SIGTERM)` 以触发已注册的监听器
  2. 等待指定的延迟时间
  3. 如果进程仍然存活则强制终止
- 如果延迟为 0 或未设置：
  1. 同步调用 `Notify(SIGTERM)`
  2. 立即终止进程

**测试支持**：`Shutdown` 函数使用内部的 `killFn` 变量（默认为操作系统的 kill），可以在测试中被替换为存根，从而在不实际终止进程的情况下测试优雅关闭行为。

## 命令执行

对外部命令的执行进行精细控制，包括超时、环境变量和生命周期管理。

```go
import (
    "context"
    "os/exec"
    "time"
    proc "go-slim.dev/proc"
)

ctx := context.Background()
err := proc.Exec(ctx, proc.ExecOptions{
    WorkDir: "/tmp",                           // 工作目录（默认为当前目录）
    Timeout: 3 * time.Second,                  // 最大执行时间
    Env:     []string{"FOO=BAR"},              // 额外的环境变量
    Command: "sh",                             // 要执行的命令
    Args:    []string{"-c", "echo ok"},        // 命令参数
    TTK:     500 * time.Millisecond,           // Time To Kill: 中断信号和强制终止信号之间的宽限期
    OnStart: func(cmd *exec.Cmd) {
        fmt.Println("Process started:", cmd.Process.Pid)
    },
})
if err != nil {
    // 处理超时、取消或执行错误
}
```

### ExecOptions 字段说明

- **WorkDir**：命令的工作目录（默认为当前进程的工作目录）
- **Timeout**：如果 > 0，会自动创建超时上下文
- **Env**：额外的环境变量（会追加到当前进程的环境变量中）
- **Stdin**、**Stdout**、**Stderr**：自定义 I/O 流（默认为 os.Stdout/os.Stderr）
- **Command**：要运行的可执行文件
- **Args**：命令行参数
- **TTK**（Time To Kill）：取消时发送中断信号和终止信号之间的延迟
- **OnStart**：命令成功启动后调用的回调函数

### 平台特定行为

- **Unix/Linux**：设置 `Setpgid=true` 创建新的进程组，防止子进程再生成子进程时出现僵尸进程
- **Windows**：不设置特殊的进程属性

## 日志控制

通过设置 `Logger` 变量控制调试输出：

```go
import (
    "io"
    "os"
    proc "go-slim.dev/proc"
)

// 禁用日志
proc.Logger = io.Discard

// 记录到文件
logFile, _ := os.Create("proc.log")
proc.Logger = logFile

// 默认：记录到 os.Stdout
```

## 使用场景

### 服务器优雅关闭

```go
proc.SetTimeToForceQuit(10 * time.Second)
proc.Once(syscall.SIGTERM, func() {
    // 关闭数据库连接
    db.Close()
    // 停止接受新请求
    server.Shutdown(context.Background())
})
```

### 带超时的构建执行

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := proc.Exec(ctx, proc.ExecOptions{
    Command: "go",
    Args:    []string{"build", "./..."},
    WorkDir: proc.WorkDir(),
    TTK:     10 * time.Second,
})
```

### 信号触发热重载

```go
proc.On(syscall.SIGHUP, func() {
    // 重新加载配置
    config.Reload()
})
```

## 开发

运行本地 CI 检查：

```bash
make ci
```

运行基准测试：

```bash
make bench
# 或保存到 artifacts/bench.txt
make bench-save
```

## 许可证

MIT