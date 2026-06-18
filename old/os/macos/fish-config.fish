if status is-interactive
    set -g fish_greeting ""

    if test -d /opt/homebrew/bin
        fish_add_path /opt/homebrew/bin
    end
    fish_add_path /usr/local/bin

    if command -q /opt/homebrew/bin/starship
        source (/opt/homebrew/bin/starship init fish --print-full-init | psub)
    else
        starship init fish | source
    end

    if command -q fnm
        fnm env --use-on-cd --shell fish | source
    end

    function set-ssh-key
        set -l key "$HOME/.ssh/$argv[1]"
        if not test -f "$key"
            echo "Key not found: $key" >&2
            echo "Available keys:" >&2
            for f in ~/.ssh/*.pub
                echo "  "(basename $f .pub) >&2
            end
            return 1
        end
        ssh-add -D 2>/dev/null
        ssh-add "$key"
        echo "Active SSH key: $argv[1]"
    end

    if test -d "$HOME/Library/pnpm"
        set -gx PNPM_HOME "$HOME/Library/pnpm"
        if not string match -q -- $PNPM_HOME $PATH
            set -gx PATH "$PNPM_HOME" $PATH
        end
    end

    zoxide init fish | source

    if command -q fzf
        fzf --fish | source
        set -gx FZF_DEFAULT_OPTS '--height 40% --layout=reverse --border'
    end
    if command -q fd
        set -gx FZF_DEFAULT_COMMAND 'fd --type f --hidden --follow --exclude .git'
        set -gx FZF_CTRL_T_COMMAND $FZF_DEFAULT_COMMAND
        set -gx FZF_ALT_C_COMMAND 'fd --type d --hidden --follow --exclude .git'
    end

    set -gx TERM xterm-256color
    set -gx COLORTERM truecolor
    set -gx EDITOR nano
    set -gx LANG zh_CN.UTF-8
    set -gx LC_ALL zh_CN.UTF-8
    set -gx LANGUAGE zh_CN.UTF-8:zh:en_US:en

    set -gx HOMEBREW_NO_AUTO_UPDATE 1

    alias ..="cd .."
    alias ...="cd ../.."
    alias l="eza --icons"
    alias la="eza -a --icons"
    alias ll="eza -lah --icons --git"
    alias ls="eza --icons"
    alias cat="bat --style=plain --paging=never"
    alias top="btop"
    alias df="duf"
    alias g="git"
    alias install="brew install"
    alias uninstall="brew uninstall"
    alias search="brew search"
    alias update="brew update"
    alias upgrade="brew upgrade"
    alias cleanup="brew cleanup --prune=all"
    alias services="brew services"
    alias cask="brew install --cask"
end
