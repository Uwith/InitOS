# InitOS

`InitOS` 以仓库根目录的 Go CLI 作为唯一主入口，提供跨平台交互式配置能力（macOS / Debian/Ubuntu，Windows 后续支持）。

## 快速开始

```bash
go run .
```

## 在 Docker（Debian）里测试（推荐）

`go.mod` 里的 `go` / `toolchain` 会要求**较新的 Go 工具链**；Debian 通过 `apt` 安装到的 `golang-go` 往往偏旧，容易出现你遇到的那种解析错误/依赖不兼容。建议直接用官方 `golang` 镜像，而不是纯 `debian` + 旧 `apt` 包。

```bash
docker run --rm -it \
  -v "/Users/你的用户名/Dev/ScriptProjects/InitOS:/workspace" \
  -w /workspace \
  golang:1.24-bookworm bash

go build ./...
go run .
```

如果你必须用 `debian:bookworm` 基础镜像，请自行安装/挂载 **Go 1.24+**（或至少 Go 1.21+ 并配置 `GOTOOLCHAIN=auto` 以自动拉取工具链），不要依赖 `apt` 的 `golang-go=1.19`。

## 交互说明

- 每次启动会先进入语言选择（会持久化到用户配置目录，并自动作为默认选中项）
- `↑/↓` 或 `j/k`：移动光标
- `Space`：勾选/取消工具
- `Enter`：生成安装队列并执行 mock 并行安装
- `/` 或直接输入：过滤工具
- `Esc`：清空过滤
- `q`：退出

## 配置来源

- 声明式工具清单：`assets/tools.yaml`
- 二进制内嵌：`go:embed`
- 结构定义：`internal/config/tools.go`
- 多语言配置：`assets/tools.yaml` 中 `default_locale`、`locales` 以及每个工具的 `name`/`desc`（可写 `en/zh/...`）
- 临时覆盖 UI 语言：环境变量 `INITOS_LOCALE`（如 `INITOS_LOCALE=zh go run .`）

## 当前内置能力

- macOS：可按需选择 `Docker`（Docker Desktop）或 `OrbStack`
- Linux（Debian/Ubuntu）：`Docker` 为镜像源与守护进程等配置；4G Swap 为独立可选项 `linux-swap`，与 Docker 无关
- Windows：暂不作为当前版本支持目标

## 目录说明

- 仓库根目录（`main.go`、`internal/`、`assets/`）：当前主线实现（Go + Bubble Tea）
- `old/`：历史脚本归档，仅用于参考，不再作为执行路径

## 构建

```bash
go build ./...
```

## 交叉编译到 Linux（给 Debian Docker 用）

在 macOS 上**一条命令**构建 `arm64` + `amd64`（产物输出到仓库根目录的 `dist/`）：

```bash
cd /path/to/InitOS
make dist
```

### 用脚本在容器里跑已编译的 Linux 二进制

`make test-docker` 等价于先 `make dist`，再执行 `scripts/docker-test.sh`。脚本会启动默认镜像 `debian:12`，将仓库根挂到容器的 `/workspace`，并把**对应当前容器 CPU 架构**的 `dist/config-cli-linux-{arm64,amd64}` 单独挂到 **`/config-cli`**，工作目录为 `/`，并执行 `/config-cli`（无额外参数，即程序的默认行为）。

```bash
cd /path/to/InitOS
make test-docker
```

**可选环境变量**（传给底层 `docker run` 前在 shell 里导出即可）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `IMAGE` | `debian:12` | 测试用基础镜像；探测架构时也会用同一镜像跑一次 `uname -m` |
| `DOCKER_TTY` | `-it` | 设为 `-i` 等可改为非交互/无 TTY（例如自动化场景） |
| `TARGET_ARCH` | 不设置 | 设为 `arm64` / `amd64`（或 `aarch64` / `x86_64`）时**跳过** `uname` 自动探测，强制选用对应 `dist` 二进制 |

强制选架构（不依赖容器内 `uname -m` 的自动检测）：

```bash
TARGET_ARCH=arm64 make test-docker
# 或
TARGET_ARCH=amd64 make test-docker
```

若你**已经**跑过 `make dist`，也可以直接执行脚本（不会再次构建）：

```bash
bash ./scripts/docker-test.sh
```

### 在容器里手动跑 `dist/` 下的文件

若你自行 `docker run` 进容器，而不是用上述脚本，请在挂载仓库后进入 `dist/`，用 **`./` 显式执行**（`PATH` 默认不包含当前目录）：

```bash
cd /workspace/dist
uname -m
chmod +x ./config-cli-linux-arm64 ./config-cli-linux-amd64
./config-cli-linux-arm64
```

- `uname -m` 为 `aarch64` / `arm64`：用 `config-cli-linux-arm64`
- `uname -m` 为 `x86_64` / `amd64`：用 `config-cli-linux-amd64`

若执行时提示 *No such file or directory* 或 *cannot execute binary file*，多半是**架构与二进制不一致**（例如在 amd64 容器里跑了 arm64 产物，或相反）。
