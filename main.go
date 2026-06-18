package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"initos/internal/config"
	"initos/internal/manager"
	"initos/internal/model"
	"initos/internal/prefs"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

//go:embed assets/tools.yaml
var toolsYAML []byte

func main() {
	app, err := config.ParseAppYAML(toolsYAML)
	if err != nil {
		log.Fatalf("load tools.yaml: %v", err)
	}

	mgr := manager.NewManager()
	tools := markInstalled(app.Tools, mgr)
	tools = onlySupported(tools)
	if len(tools) == 0 {
		log.Fatalf("no installable items for this platform (check install commands in tools.yaml)")
	}

	activeLocale := os.Getenv("INITOS_LOCALE")
	if activeLocale == "" {
		var perr error
		activeLocale, perr = prefs.ReadLocale()
		if perr != nil {
			log.Fatalf("read saved locale: %v", perr)
		}
	}
	preferredLocale := app.DefaultLocale
	if activeLocale != "" && knownLocaleID(activeLocale, app.Locales) {
		preferredLocale = activeLocale
	}

	p := tea.NewProgram(
		model.New(tools, mgr, model.UIOptions{
			AppDefaultLocale: preferredLocale,
			Locales:          app.Locales,
			ActiveUILocale:   "",
		}),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI: %v", err)
	}
}

func knownLocaleID(id string, list []config.LocaleEntry) bool {
	for _, e := range list {
		if e.ID == id {
			return true
		}
	}
	return false
}

func markInstalled(tools []config.Tool, mgr manager.Manager) []config.Tool {
	for i := range tools {
		command, err := mgr.CommandFor(tools[i])
		if err != nil {
			fmt.Printf("系统能力检查失败（%s）: %v\n", tools[i].ID, err)
			tools[i].Supported = false
			continue
		}
		tools[i].Supported = command != ""
		if !tools[i].Supported {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		installed, err := mgr.IsInstalled(ctx, tools[i])
		cancel()
		if err != nil {
			fmt.Printf("预检查失败（%s）: %v\n", tools[i].ID, err)
			continue
		}
		tools[i].Installed = installed
	}
	return tools
}

// onlySupported 只保留本机存在安装命令的工具，使各系统看到的可装列表不同（例如 Windows/Linux 不显示 OrbStack）。
func onlySupported(tools []config.Tool) []config.Tool {
	out := make([]config.Tool, 0, len(tools))
	for i := range tools {
		if tools[i].Supported {
			out = append(out, tools[i])
		}
	}
	return out
}
