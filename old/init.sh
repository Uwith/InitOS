#!/usr/bin/env bash
# =============================================================================
# 统一初始化入口：按当前 OS 选择并执行对应脚本
#   bash init.sh
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GRN='\033[0;32m'
NC='\033[0m'

log_info()  { echo -e "${GRN}[init]${NC} $*"; }
log_error() { echo -e "${RED}[init]${NC} $*" >&2; }

select_run_mode() {
  local answer=""
  echo "请选择执行模式：" >&2
  echo "  1) 手动 manual（默认，关键步骤会确认）" >&2
  echo "  2) 自动 auto  （跳过确认）" >&2
  read -r -p "输入选项 [1/2]（默认 1）: " answer
  case "$answer" in
    2)
      echo "auto"
      ;;
    ""|1)
      echo "manual"
      ;;
    *)
      log_info "无效选项，使用默认手动模式。"
      echo "manual"
      ;;
  esac
}

confirm_continue() {
  local prompt="$1"
  local answer=""
  read -r -p "$prompt [y/N]: " answer
  case "$answer" in
    y|Y|yes|YES)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

run_script() {
  local script_path="$1"
  shift

  if [[ ! -f "$script_path" ]]; then
    log_error "未找到脚本：$script_path"
    exit 1
  fi

  log_info "执行：$script_path"
  exec bash "$script_path" "$@"
}

detect_linux_family() {
  if [[ -f /etc/os-release ]]; then
    local id=""
    local id_like=""
    # shellcheck disable=SC1091
    source /etc/os-release
    id="${ID:-}"
    id_like="${ID_LIKE:-}"
    if [[ "$id" == "debian" || "$id" == "ubuntu" || "$id_like" == *"debian"* ]]; then
      echo "debian"
      return 0
    fi
  fi
  echo "unknown"
}

main() {
  local run_mode=""
  local os_name
  local os_desc=""
  local target_script=""
  local linux_family=""

  run_mode="$(select_run_mode)"
  log_info "当前模式：$run_mode"

  os_name="$(uname -s)"

  case "$os_name" in
    Darwin)
      os_desc="macOS (Darwin)"
      target_script="$SCRIPT_DIR/init-macos.sh"
      ;;
    Linux)
      linux_family="$(detect_linux_family)"
      case "$linux_family" in
        debian)
          os_desc="Linux (Debian/Ubuntu family)"
          target_script="$SCRIPT_DIR/init-debian-server.sh"
          ;;
        *)
          os_desc="Linux (unsupported family)"
          log_info "操作系统识别结果：$os_desc"
          log_error "当前 Linux 发行版不在支持范围（仅支持 Debian/Ubuntu）。"
          exit 1
          ;;
      esac
      ;;
    *)
      os_desc="$os_name (unsupported)"
      log_info "操作系统识别结果：$os_desc"
      log_error "不支持的系统：$os_name"
      exit 1
      ;;
  esac

  log_info "操作系统识别结果：$os_desc"
  if [[ "$run_mode" == "manual" ]]; then
    if ! confirm_continue "确认继续执行初始化脚本吗？"; then
      log_info "已取消执行。"
      exit 0
    fi
  fi

  run_script "$target_script" "$@"
}

main "$@"
