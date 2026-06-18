package model

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"initos/internal/config"
	"initos/internal/locale"
	"initos/internal/manager"
	"initos/internal/prefs"
	"initos/internal/ui"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type viewMode int

const (
	langPickMode viewMode = iota
	browseMode
	installMode
	doneMode
)

// UIOptions 语言与首屏行为：ActiveUILocale 为空时先进入语言选择；非空则直接进入浏览（并用于工具文案）。
type UIOptions struct {
	AppDefaultLocale string
	Locales          []config.LocaleEntry
	ActiveUILocale   string
}

type installStage string

const (
	stagePending installStage = "pending"
	stageRunning installStage = "running"
	stageDone    installStage = "done"
	stageFailed  installStage = "failed"
)

type installResultMsg struct {
	ToolID string
	Stage  installStage
	Err    error
}

type installClosedMsg struct{}

type Model struct {
	tools      []config.Tool
	manager    manager.Manager
	uiOpts     UIOptions
	uiLang     string
	langCursor int
	theme      ui.Theme
	filter     textinput.Model
	progress   progress.Model
	spinner    spinner.Model
	mode       viewMode
	cursor     int
	selected   map[string]bool
	width      int
	height     int
	lastNotice string

	installStates    map[string]installStage
	installQueue     []config.Tool
	installQueueByID map[string]config.Tool
	installDone      int
	installFailed    int
	workers          int
	installCh        <-chan installResultMsg
}

func New(tools []config.Tool, mgr manager.Manager, opts UIOptions) Model {
	filter := textinput.New()
	filter.CharLimit = 128
	filter.Width = 42
	filter.Focus()

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 42

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.NewTheme().Selected

	m := Model{
		tools:            tools,
		manager:          mgr,
		uiOpts:           opts,
		theme:            ui.NewTheme(),
		filter:           filter,
		progress:         p,
		spinner:          s,
		selected:         make(map[string]bool),
		installStates:    make(map[string]installStage),
		installQueueByID: make(map[string]config.Tool),
		workers:          3,
	}
	if opts.ActiveUILocale != "" {
		m.uiLang = opts.ActiveUILocale
		m.mode = browseMode
		m.applyFilterI18n()
	} else {
		m.uiLang = opts.AppDefaultLocale
		if m.uiLang == "" {
			m.uiLang = locale.EN
		}
		m.mode = langPickMode
		m.langCursor = indexOfLocale(opts.Locales, opts.AppDefaultLocale)
	}
	return m
}

func indexOfLocale(list []config.LocaleEntry, id string) int {
	for i, e := range list {
		if e.ID == id {
			return i
		}
	}
	return 0
}

func (m *Model) applyFilterI18n() {
	m.filter.Placeholder = locale.UIText(m.uiLang, "filter.placeholder")
	m.filter.Prompt = locale.UIText(m.uiLang, "filter.tui_prompt")
}

func (m Model) tr(key string) string {
	return locale.UIText(m.uiLang, key)
}

func (m Model) toolName(t config.Tool) string {
	return t.TextName(m.uiLang)
}

func (m Model) toolDesc(t config.Tool) string {
	return t.TextDesc(m.uiLang)
}

func (m Model) Init() tea.Cmd {
	if m.mode == browseMode {
		return textinput.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if msg.Width > 20 {
			m.filter.Width = min(msg.Width-18, 60)
			m.progress.Width = min(msg.Width-20, 60)
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case langPickMode:
			return m.updateLangPick(msg)
		case browseMode:
			return m.updateBrowse(msg)
		case installMode:
			return m.updateInstallKeys(msg)
		case doneMode:
			return m.updateDoneKeys(msg)
		}

	case spinner.TickMsg:
		if m.mode == installMode {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case installResultMsg:
		switch msg.Stage {
		case stageRunning:
			m.installStates[msg.ToolID] = stageRunning
		case stageDone:
			m.installStates[msg.ToolID] = stageDone
			m.installDone++
			_ = m.progress.SetPercent(float64(m.installDone+m.installFailed) / float64(len(m.installQueue)))
		case stageFailed:
			m.installStates[msg.ToolID] = stageFailed
			m.installFailed++
			_ = m.progress.SetPercent(float64(m.installDone+m.installFailed) / float64(len(m.installQueue)))
			tname := m.installQueueByID[msg.ToolID].TextName(m.uiLang)
			m.lastNotice = fmt.Sprintf(m.tr("notify.fail"), tname, msg.Err)
		}
		return m, waitInstallMsg(m.installCh)

	case installClosedMsg:
		m.mode = doneMode
		if m.installFailed == 0 {
			m.lastNotice = fmt.Sprintf(m.tr("notify.all_ok"), m.installDone)
		} else {
			m.lastNotice = fmt.Sprintf(m.tr("notify.partial"), m.installDone, m.installFailed)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case langPickMode:
		return m.viewLanguage()
	case browseMode:
		return m.viewBrowse()
	case installMode:
		return m.viewInstall()
	case doneMode:
		return m.viewDone()
	default:
		return ""
	}
}

func (m Model) updateLangPick(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.langCursor > 0 {
			m.langCursor--
		}
		return m, nil
	case "down", "j":
		if m.langCursor < len(m.uiOpts.Locales)-1 {
			m.langCursor++
		}
		return m, nil
	case "enter":
		if len(m.uiOpts.Locales) == 0 {
			m.uiLang = locale.EN
		} else {
			m.uiLang = m.uiOpts.Locales[m.langCursor].ID
		}
		if m.uiLang == "" {
			m.uiLang = m.uiOpts.AppDefaultLocale
		}
		if m.uiLang == "" {
			m.uiLang = locale.EN
		}
		m.applyFilterI18n()
		if err := prefs.WriteLocale(m.uiLang); err != nil {
			m.lastNotice = fmt.Sprintf(m.tr("notice.locale_save_failed"), err)
		}
		m.mode = browseMode
		m.cursor = 0
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		maxCursor := len(m.filteredIndices()) - 1
		if maxCursor >= 0 && m.cursor < maxCursor {
			m.cursor++
		}
		return m, nil
	case " ":
		indices := m.filteredIndices()
		if len(indices) == 0 {
			return m, nil
		}
		if m.cursor >= len(indices) {
			m.cursor = len(indices) - 1
		}
		tool := m.tools[indices[m.cursor]]
		if tool.Installed {
			m.lastNotice = fmt.Sprintf(m.tr("notice.installed"), m.toolName(tool))
			return m, nil
		}
		m.selected[tool.ID] = !m.selected[tool.ID]
		return m, nil
	case "enter":
		queue := m.buildInstallQueue()
		if len(queue) == 0 {
			m.lastNotice = m.tr("notice.select_first")
			return m, nil
		}
		m.mode = installMode
		m.installQueue = queue
		m.installDone = 0
		m.installFailed = 0
		m.installStates = make(map[string]installStage, len(queue))
		m.installQueueByID = make(map[string]config.Tool, len(queue))
		for _, t := range queue {
			m.installStates[t.ID] = stagePending
			m.installQueueByID[t.ID] = t
		}
		_ = m.progress.SetPercent(0)
		m.installCh = startWorkerPool(queue, m.workers, m.manager)
		return m, tea.Batch(waitInstallMsg(m.installCh), m.spinner.Tick)
	case "esc", "ctrl+[":
		m.filter.SetValue("")
		m.cursor = 0
		return m, nil
	case "/":
		m.filter.SetValue("")
		m.lastNotice = m.tr("notice.filter_start")
		return m, nil
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	if m.cursor > len(m.filteredIndices())-1 {
		m.cursor = max(0, len(m.filteredIndices())-1)
	}
	return m, cmd
}

func (m Model) updateInstallKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateDoneKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		for i := range m.tools {
			if m.installStates[m.tools[i].ID] == stageDone {
				m.tools[i].Installed = true
			}
		}
		m.mode = browseMode
		m.installQueue = nil
		m.installStates = map[string]installStage{}
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m Model) viewLanguage() string {
	header := m.theme.Header.Render(ui.GradientText(m.tr("lang.title")))
	help := m.theme.Help.Render(m.tr("lang.hint"))
	if len(m.uiOpts.Locales) == 0 {
		return strings.Join([]string{
			header,
			"",
			m.theme.Subtle.Render(m.tr("lang.empty")),
			help,
		}, "\n")
	}
	lines := make([]string, 0, len(m.uiOpts.Locales))
	for i, e := range m.uiOpts.Locales {
		cursor := "  "
		if i == m.langCursor {
			cursor = m.theme.Cursor.Render("➜ ")
		}
		label := fmt.Sprintf("%s [%s] %s", cursor, e.ID, e.Label)
		if i == m.langCursor {
			label = m.theme.Normal.Bold(true).Render(label)
		}
		lines = append(lines, label)
	}
	return strings.Join([]string{
		header,
		"",
		strings.Join(lines, "\n"),
		"",
		help,
	}, "\n")
}

func (m Model) viewBrowse() string {
	header := m.theme.Header.Render(ui.GradientText(m.tr("app.subtitle")))
	help := m.theme.Help.Render(m.tr("help.browse"))
	if len(m.tools) == 0 {
		return strings.Join([]string{
			header,
			"",
			m.theme.Subtle.Render(m.tr("empty.tools")),
			help,
		}, "\n")
	}
	filterLine := m.theme.FilterPrompt.Render(m.tr("filter.label")) + m.filter.View()

	indices := m.filteredIndices()
	if len(indices) == 0 {
		return strings.Join([]string{
			header,
			filterLine,
			m.theme.Subtle.Render(m.tr("empty.filter")),
			help,
		}, "\n")
	}

	maxNameW := 0
	for _, idx := range indices {
		if w := runewidth.StringWidth(m.toolName(m.tools[idx])); w > maxNameW {
			maxNameW = w
		}
	}

	lines := make([]string, 0, len(indices))
	for pos, idx := range indices {
		tool := m.tools[idx]
		cursor := "  "
		if pos == m.cursor {
			cursor = m.theme.Cursor.Render("➜ ")
		}
		checked := "[ ]"
		if m.selected[tool.ID] {
			checked = m.theme.Selected.Render("[x]")
		}
		status := ""
		if tool.Installed {
			status = " " + m.theme.Installed.Render(m.tr("status.installed"))
		}
		namePadded := alignNameColumn(m.toolName(tool), maxNameW)
		label := fmt.Sprintf("%s %s %s - %s%s", cursor, checked, namePadded, m.toolDesc(tool), status)
		if pos == m.cursor {
			label = m.theme.Normal.Bold(true).Render(label)
		}
		lines = append(lines, label)
	}

	footer := fmt.Sprintf(m.tr("footer.selected"), m.selectedCount())
	if m.lastNotice != "" {
		footer = footer + "\n" + m.theme.Subtle.Render(m.lastNotice)
	}

	return strings.Join([]string{
		header,
		filterLine,
		"",
		strings.Join(lines, "\n"),
		"",
		m.theme.Subtle.Render(footer),
		help,
	}, "\n")
}

func (m Model) viewInstall() string {
	title := m.theme.Header.Render(ui.GradientText(m.tr("install.title")))
	progressLine := m.progress.View()
	spinLine := fmt.Sprintf(m.tr("install.sub"), m.spinner.View(), min(m.workers, len(m.installQueue)))

	lines := make([]string, 0, len(m.installQueue))
	for _, tool := range m.installQueue {
		stage := m.installStates[tool.ID]
		status := stageLabel(stage, m.uiLang)
		lines = append(lines, fmt.Sprintf("- %s: %s", m.toolName(tool), status))
	}

	help := m.theme.Help.Render(m.tr("install.help"))
	return strings.Join([]string{
		title,
		progressLine,
		spinLine,
		"",
		strings.Join(lines, "\n"),
		"",
		help,
	}, "\n")
}

func (m Model) viewDone() string {
	title := m.theme.Header.Render(ui.GradientText(m.tr("done.title")))
	outcome := m.theme.Success.Render(fmt.Sprintf(m.tr("done.ok"), m.installDone))
	if m.installFailed > 0 {
		outcome += "  " + m.theme.Failure.Render(fmt.Sprintf(m.tr("done.fail"), m.installFailed))
	}
	lines := []string{title, outcome}
	if m.lastNotice != "" {
		lines = append(lines, m.theme.Subtle.Render(m.lastNotice))
	}
	lines = append(lines, m.theme.Help.Render(m.tr("done.hint")))
	return strings.Join(lines, "\n")
}

func (m Model) filteredIndices() []int {
	kw := strings.TrimSpace(strings.ToLower(m.filter.Value()))
	indices := make([]int, 0, len(m.tools))
	for idx, tool := range m.tools {
		if kw == "" {
			indices = append(indices, idx)
			continue
		}
		search := strings.ToLower(tool.SearchText())
		if strings.Contains(search, kw) {
			indices = append(indices, idx)
		}
	}
	return indices
}

func (m Model) selectedCount() int {
	total := 0
	for _, ok := range m.selected {
		if ok {
			total++
		}
	}
	return total
}

func (m Model) buildInstallQueue() []config.Tool {
	queue := make([]config.Tool, 0, len(m.selected))
	for _, t := range m.tools {
		if m.selected[t.ID] && !t.Installed && t.Supported {
			queue = append(queue, t)
		}
	}
	sort.Slice(queue, func(i, j int) bool { return queue[i].ID < queue[j].ID })
	return queue
}

func waitInstallMsg(ch <-chan installResultMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return installClosedMsg{}
		}
		return msg
	}
}

func startWorkerPool(queue []config.Tool, workers int, mgr manager.Manager) <-chan installResultMsg {
	out := make(chan installResultMsg)
	if len(queue) == 0 {
		close(out)
		return out
	}

	workerCount := min(workers, len(queue))
	jobs := make(chan config.Tool)
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tool := range jobs {
				out <- installResultMsg{ToolID: tool.ID, Stage: stageRunning}
				if err := mgr.Install(ctx, tool); err != nil {
					out <- installResultMsg{ToolID: tool.ID, Stage: stageFailed, Err: err}
					continue
				}
				out <- installResultMsg{ToolID: tool.ID, Stage: stageDone}
			}
		}()
	}

	go func() {
		for _, tool := range queue {
			jobs <- tool
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()
	return out
}

func stageLabel(stage installStage, lang string) string {
	switch stage {
	case stagePending:
		return locale.UIText(lang, "stage.pending")
	case stageRunning:
		return locale.UIText(lang, "stage.running")
	case stageDone:
		return locale.UIText(lang, "stage.done")
	case stageFailed:
		return locale.UIText(lang, "stage.failed")
	default:
		return locale.UIText(lang, "stage.unknown")
	}
}

// alignNameColumn 将名称用空格垫到 displayWidth，使多行「名称 - 描述」中的「 - 」列对齐。
func alignNameColumn(name string, displayWidth int) string {
	w := runewidth.StringWidth(name)
	if w >= displayWidth {
		return name
	}
	return name + strings.Repeat(" ", displayWidth-w)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
