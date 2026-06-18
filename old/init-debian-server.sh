#!/usr/bin/env bash
# =============================================================================
# 单文件服务器初始化：内含本机 Starship + Fish 习惯，拷到任意 Debian 系机器
#   bash init-debian-server.sh
# 需 root（或由具备 root 的环境执行）。完成后重登 SSH。
# =============================================================================
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a
export UCF_FORCE_CONFFOLD=1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GRN='\033[0;32m'
YLW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GRN}[init]${NC} $*"; }
log_warn()  { echo -e "${YLW}[init]${NC} $*"; }
log_error() { echo -e "${RED}[init]${NC} $*" >&2; }

apt_noninteractive() {
  apt-get -y \
    -o Dpkg::Options::=--force-confdef \
    -o Dpkg::Options::=--force-confold \
    "$@"
}

require_root() {
  if [[ "${EUID:-}" -ne 0 ]]; then
    log_error "请使用 root 运行：sudo bash $0"
    exit 1
  fi
}

resolve_target_user() {
  TARGET_USER="root"
  TARGET_HOME="/root"
}

run_as_target() {
  "$@"
}

setup_swap_2g() {
  if swapon --show 2>/dev/null | grep -qF '/swapfile'; then
    log_info "交换分区已启用 /swapfile，跳过创建。"
  else
    if [[ ! -f /swapfile ]]; then
      log_info "创建 2G /swapfile …"
      if ! fallocate -l 2G /swapfile 2>/dev/null; then
        log_warn "fallocate 失败，改用 dd（较慢）…"
        dd if=/dev/zero of=/swapfile bs=1M count=2048 status=progress
      fi
      chmod 600 /swapfile
      mkswap /swapfile
    fi
    swapon /swapfile
  fi
  if ! grep -qE '^\s*/swapfile\s' /etc/fstab 2>/dev/null; then
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    log_info "已写入 /etc/fstab（/swapfile）。"
  fi
}

setup_locales() {
  log_info "配置 locale（en_US.UTF-8 + zh_CN.UTF-8，系统默认 en_US）…"
  apt_noninteractive install locales
  sed -i 's/^# *\(en_US.UTF-8 UTF-8\)/\1/' /etc/locale.gen
  sed -i 's/^# *\(zh_CN.UTF-8 UTF-8\)/\1/' /etc/locale.gen
  locale-gen
  update-locale LANG=en_US.UTF-8 LC_CTYPE=en_US.UTF-8
}

setup_packages() {
  log_info "apt update / upgrade …"
  apt-get update
  apt_noninteractive upgrade
  log_info "安装 Fish / Starship / 现代 CLI …"
  apt_noninteractive install \
    fish starship eza bat btop duf \
    fd-find ripgrep fzf zoxide curl git
}

install_user_configs() {
  local fish_dir="$TARGET_HOME/.config/fish"
  local starship_path="$TARGET_HOME/.config/starship.toml"
  local starship_template="$SCRIPT_DIR/config/starship.toml"
  local fish_template="$SCRIPT_DIR/os/debian/fish-config.fish"
  log_info "为 $TARGET_USER 写入 Fish / Starship / Ghostty terminfo …"

  run_as_target mkdir -p "$fish_dir" "$TARGET_HOME/.terminfo/x" "$TARGET_HOME/.config"

  local xterm256=""
  for p in /lib/terminfo/x/xterm-256color /usr/share/terminfo/x/xterm-256color; do
    if [[ -e "$p" ]]; then
      xterm256="$p"
      break
    fi
  done
  if [[ -n "$xterm256" ]]; then
    run_as_target ln -sf "$xterm256" "$TARGET_HOME/.terminfo/x/xterm-ghostty"
  else
    log_warn "未找到 xterm-256color terminfo，跳过 xterm-ghostty 软链。"
  fi

  if [[ ! -f "$starship_template" ]]; then
    log_error "未找到 Starship 模板文件：$starship_template"
    exit 1
  fi
  log_info "写入 $starship_path （来源：$starship_template）…"
  run_as_target install -m 0644 "$starship_template" "$starship_path"

  if [[ ! -f "$fish_template" ]]; then
    log_error "未找到 Fish 模板文件：$fish_template"
    exit 1
  fi
  log_info "写入 $fish_dir/config.fish （来源：$fish_template）…"
  run_as_target install -m 0644 "$fish_template" "$fish_dir/config.fish"

  local fish_bin
  fish_bin="$(command -v fish)"
  if ! grep -qF "$fish_bin" /etc/shells 2>/dev/null; then
    echo "$fish_bin" >> /etc/shells
  fi
  chsh -s "$fish_bin" "$TARGET_USER"
  log_info "已将 $TARGET_USER 的默认 shell 设为 $fish_bin。请重新登录 SSH 后生效。"
}

main() {
  require_root
  resolve_target_user
  log_info "目标用户：$TARGET_USER ($TARGET_HOME)"

  setup_swap_2g
  setup_locales
  setup_packages
  install_user_configs

  log_info "完成。建议：重新登录后运行 mem、ll、top（btop）自检。"
}

main "$@"
