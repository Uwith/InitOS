package locale

// Code 为 UI 与 tools.yaml 中 locales[].id 使用的语言码（如 en、zh）。
const (
	EN = "en"
	ZH = "zh"
)

// UIText 返回 TUI 固定文案。未翻译的 key 回退为英文，再无则 [key]。
func UIText(lang, key string) string {
	langM := uiMessages[lang]
	if langM == nil {
		langM = uiMessages[EN]
	}
	if s := langM[key]; s != "" {
		return s
	}
	if s := uiMessages[EN][key]; s != "" {
		return s
	}
	return "[" + key + "]"
}

var uiMessages = map[string]map[string]string{
	EN: {
		"app.subtitle":              "InitOS CLI · Cross-platform setup",
		"lang.title":                "Choose language",
		"lang.hint":                 "↑/↓ or j/k · Enter confirm · q quit",
		"lang.empty":                "No locale options configured.",
		"filter.label":              "Filter: ",
		"filter.tui_prompt":         "filter> ",
		"filter.placeholder":        "filter by name / id / desc",
		"help.browse":               "↑/↓ or j/k · space toggle · Enter run · / filter · Esc clear · q quit",
		"empty.tools":               "No tools to show for this system.",
		"empty.filter":              "No matches",
		"footer.selected":           "Selected: %d (installed skipped)",
		"notice.installed":          "%s is already installed, toggle skipped",
		"notice.select_first":       "Select at least one not-yet-installed tool",
		"notice.filter_start":       "Type to filter",
		"notice.locale_save_failed": "Failed to save locale preference: %v",
		"install.title":             "Installing",
		"install.sub":               "%s parallel (workers=%d)",
		"install.help":              "q to quit (mock install; does not stop real commands)",
		"done.title":                "Install Finished",
		"done.ok":                   "OK: %d",
		"done.fail":                 "Failed: %d",
		"done.hint":                 "Enter: back to list · q: quit",
		"stage.pending":             "Pending",
		"stage.running":             "Running",
		"stage.done":                "Done",
		"stage.failed":              "Failed",
		"stage.unknown":             "Unknown",
		"status.installed":          "(installed)",
		"notify.all_ok":             "All %d install task(s) succeeded",
		"notify.partial":            "Done: %d ok, %d failed",
		"notify.fail":               "Install failed: %s (%v)",
	},
	ZH: {
		"app.subtitle":              "InitOS CLI · 跨平台配置",
		"lang.title":                "选择语言",
		"lang.hint":                 "↑/↓ 或 j/k · Enter 确定 · q 退出",
		"lang.empty":                "未配置可选语言。",
		"filter.label":              "过滤: ",
		"filter.tui_prompt":         "filter> ",
		"filter.placeholder":        "按名称 / id / 描述 过滤",
		"help.browse":               "↑/↓ 或 j/k · 空格 选择 · Enter 执行 · / 过滤 · Esc 清空 · q 退出",
		"empty.tools":               "当前没有可显示的工具。",
		"empty.filter":              "没有匹配项",
		"footer.selected":           "已选择 %d 项（已安装项将自动跳过）",
		"notice.installed":          "%s 已安装，跳过选择",
		"notice.select_first":       "请先选择至少一个未安装工具",
		"notice.filter_start":       "开始输入以过滤",
		"notice.locale_save_failed": "保存语言偏好失败：%v",
		"install.title":             "正在安装",
		"install.sub":               "%s 并行执行中（worker=%d）",
		"install.help":              "安装中可按 q 退出（本版本为 mock）",
		"done.title":                "安装结束",
		"done.ok":                   "成功: %d",
		"done.fail":                 "失败: %d",
		"done.hint":                 "按 Enter 返回列表，或按 q 退出",
		"stage.pending":             "等待中",
		"stage.running":             "进行中",
		"stage.done":                "已完成",
		"stage.failed":              "失败",
		"stage.unknown":             "未知",
		"status.installed":          "(已安装)",
		"notify.all_ok":             "全部完成：%d 个工具安装任务成功",
		"notify.partial":            "任务完成：成功 %d，失败 %d",
		"notify.fail":               "安装失败：%s (%v)",
	},
}

// SupportedUICodes 内置 UI 已有翻译的语言码，用于与 tools.yaml 的 locales 交叉校验时提示。
func SupportedUICodes() []string { return []string{EN, ZH} }
