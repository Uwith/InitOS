#!/usr/bin/env bash
# =============================================================================
# macOS 初始化脚本：安装 Homebrew + Fish + Starship + 现代 CLI 工具
#   bash init-macos.sh
# 无需 root，以普通用户身份运行即可。完成后重新打开终端。
# =============================================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GRN='\033[0;32m'
YLW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GRN}[init]${NC} $*"; }
log_warn()  { echo -e "${YLW}[init]${NC} $*"; }
log_error() { echo -e "${RED}[init]${NC} $*" >&2; }

TARGET_USER="$(whoami)"
TARGET_HOME="$HOME"

# ── Xcode Command Line Tools ────────────────────────────────────────────────
install_xcode_cli() {
  if xcode-select -p &>/dev/null; then
    log_info "Xcode Command Line Tools 已安装，跳过。"
  else
    log_info "安装 Xcode Command Line Tools …"
    xcode-select --install 2>/dev/null || true
    log_warn "请在弹窗中确认安装，完成后重新运行此脚本。"
    exit 0
  fi
}

# ── Homebrew ─────────────────────────────────────────────────────────────────
install_homebrew() {
  if command -v brew &>/dev/null; then
    log_info "Homebrew 已安装，跳过。"
  else
    log_info "安装 Homebrew …"
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  fi

  if [[ -x /opt/homebrew/bin/brew ]]; then
    eval "$(/opt/homebrew/bin/brew shellenv)"
  elif [[ -x /usr/local/bin/brew ]]; then
    eval "$(/usr/local/bin/brew shellenv)"
  fi

  brew update
}

# ── Brew 包 ──────────────────────────────────────────────────────────────────
setup_packages() {
  log_info "安装 Fish / Starship / 现代 CLI …"

  local pkgs=(
    fish
    starship
    eza
    bat
    btop
    duf
    fd
    ripgrep
    fzf
    zoxide
    curl
    git
    fnm
  )

  for pkg in "${pkgs[@]}"; do
    if brew list --formula "$pkg" &>/dev/null; then
      log_info "$pkg 已安装，跳过。"
    else
      brew install "$pkg"
    fi
  done
}

# ── 用户配置 ─────────────────────────────────────────────────────────────────
install_user_configs() {
  local fish_dir="$TARGET_HOME/.config/fish"
  local starship_path="$TARGET_HOME/.config/starship.toml"
  local starship_template="$SCRIPT_DIR/config/starship.toml"
  local fish_template="$SCRIPT_DIR/os/macos/fish-config.fish"
  log_info "为 $TARGET_USER 写入 Fish / Starship 配置 …"

  mkdir -p "$fish_dir" "$TARGET_HOME/.config"

  if [[ ! -f "$starship_template" ]]; then
    log_error "未找到 Starship 模板文件：$starship_template"
    exit 1
  fi
  log_info "写入 $starship_path （来源：$starship_template）…"
  install -m 0644 "$starship_template" "$starship_path"

  if [[ ! -f "$fish_template" ]]; then
    log_error "未找到 Fish 模板文件：$fish_template"
    exit 1
  fi
  log_info "写入 $fish_dir/config.fish （来源：$fish_template）…"
  install -m 0644 "$fish_template" "$fish_dir/config.fish"
}

# ── 设置默认 Shell ──────────────────────────────────────────────────────────
set_default_shell() {
  local fish_bin
  fish_bin="$(command -v fish)"

  if [[ -z "$fish_bin" ]]; then
    log_error "未找到 fish，无法设置默认 shell。"
    return 1
  fi

  if ! grep -qF "$fish_bin" /etc/shells 2>/dev/null; then
    log_info "将 $fish_bin 添加到 /etc/shells（需要 sudo）…"
    echo "$fish_bin" | sudo tee -a /etc/shells >/dev/null
  fi

  if [[ "$SHELL" == "$fish_bin" ]]; then
    log_info "默认 shell 已是 Fish，跳过。"
  else
    log_info "设置默认 shell 为 $fish_bin …"
    chsh -s "$fish_bin"
  fi

  log_info "已将 ${TARGET_USER} 的默认 shell 设为 ${fish_bin}。请重新打开终端后生效。"
}

# ── macOS 系统偏好 ──────────────────────────────────────────────────────────
setup_macos_defaults() {
  log_info "调整 macOS 系统偏好 …"

  # Finder: 显示隐藏文件
  defaults write com.apple.finder AppleShowAllFiles -bool true

  # Finder: 显示文件扩展名
  defaults write NSGlobalDomain AppleShowAllExtensions -bool true

  # Finder: 显示路径栏
  defaults write com.apple.finder ShowPathbar -bool true

  # Finder: 显示状态栏
  defaults write com.apple.finder ShowStatusBar -bool true

  # Finder: 默认以列表视图打开
  defaults write com.apple.finder FXPreferredViewStyle -string "Nlsv"

  # Finder: 搜索时默认搜索当前目录
  defaults write com.apple.finder FXDefaultSearchScope -string "SCcf"

  # 禁用 .DS_Store 在网络和 USB 卷上
  defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true
  defaults write com.apple.desktopservices DSDontWriteUSBStores -bool true

  # Dock: 自动隐藏
  defaults write com.apple.dock autohide -bool true

  # Dock: 缩小图标尺寸
  defaults write com.apple.dock tilesize -int 48

  # 键盘：加快按键重复速率
  defaults write NSGlobalDomain KeyRepeat -int 2
  defaults write NSGlobalDomain InitialKeyRepeat -int 15

  # 截图保存到 ~/Pictures/Screenshots
  local ss_dir="$HOME/Pictures/Screenshots"
  mkdir -p "$ss_dir"
  defaults write com.apple.screencapture location -string "$ss_dir"

  # 重启 Finder 和 Dock 以应用更改
  killall Finder 2>/dev/null || true
  killall Dock 2>/dev/null || true

  # 禁用终端 "Last login" 消息
  touch "$HOME/.hushlogin"

  log_info "macOS 系统偏好已调整。"
}

# ── 主函数 ───────────────────────────────────────────────────────────────────
main() {
  if [[ "$(uname)" != "Darwin" ]]; then
    log_error "此脚本仅适用于 macOS。"
    exit 1
  fi

  log_info "当前用户：$TARGET_USER ($TARGET_HOME)"

  install_xcode_cli
  install_homebrew
  setup_packages
  install_user_configs
  set_default_shell
  setup_macos_defaults

  log_info "完成！请重新打开终端，运行 ll、top（btop）自检。"
}

main "$@"
