package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LocaleEntry 语言列表项（用于首屏选择器）；label 用该语言自称即可（如 English / 中文）。
type LocaleEntry struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
}

// App 根配置（tools.yaml 顶层除 tools 外可含 default_locale 与 locales）。
type App struct {
	DefaultLocale string        `yaml:"default_locale"`
	Locales       []LocaleEntry `yaml:"locales"`
	Tools         []Tool        `yaml:"tools"`
}

// ParseAppYAML 解析完整应用配置。
func ParseAppYAML(content []byte) (*App, error) {
	var root struct {
		DefaultLocale string        `yaml:"default_locale"`
		Locales       []LocaleEntry `yaml:"locales"`
		Tools         []Tool        `yaml:"tools"`
	}
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse app yaml: %w", err)
	}
	if len(root.Tools) == 0 {
		return nil, fmt.Errorf("no tools found in config")
	}
	if root.DefaultLocale == "" {
		root.DefaultLocale = "en"
	}
	if len(root.Locales) == 0 {
		root.Locales = defaultLocaleEntries()
	}
	for i := range root.Tools {
		t := &root.Tools[i]
		if t.ID == "" {
			return nil, fmt.Errorf("tool at index %d must have id", i)
		}
		if !t.Name.HasText() {
			return nil, fmt.Errorf("tool %q must have a non-empty name (string or i18n map)", t.ID)
		}
	}
	return &App{
		DefaultLocale: root.DefaultLocale,
		Locales:       root.Locales,
		Tools:         root.Tools,
	}, nil
}

func defaultLocaleEntries() []LocaleEntry {
	return []LocaleEntry{
		{ID: "en", Label: "English"},
		{ID: "zh", Label: "中文"},
	}
}

// Tool 展示名/描述在指定语言下解析（缺省同 LocalizedString.Text）。
func (t Tool) TextName(code string) string { return t.Name.Text(code) }

// TextDesc 工具描述（当前语言 + 回退）。
func (t Tool) TextDesc(code string) string { return t.Desc.Text(code) }

// SearchText 过滤：合并 id 与各语言 name/desc。
func (t Tool) SearchText() string {
	b := t.ID
	for _, s := range t.Name.AllValues() {
		if s != "" {
			b += " " + s
		}
	}
	for _, s := range t.Desc.AllValues() {
		if s != "" {
			b += " " + s
		}
	}
	return b
}
