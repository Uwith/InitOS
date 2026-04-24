package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LocalizedString 在 YAML 中可写为单字符串（视为 en）或多语言表（如 en/zh/ja）。
type LocalizedString struct {
	byLocale map[string]string
}

// UnmarshalYAML 支持 `name: Git` 或 `name: { en: "Git", zh: "..." }`。
func (l *LocalizedString) UnmarshalYAML(n *yaml.Node) error {
	if n == nil {
		return fmt.Errorf("localized: nil node")
	}
	if l.byLocale == nil {
		l.byLocale = make(map[string]string)
	}
	switch n.Kind {
	case yaml.ScalarNode:
		var s string
		if err := n.Decode(&s); err != nil {
			return err
		}
		l.byLocale["en"] = s
		return nil
	case yaml.MappingNode:
		var m map[string]string
		if err := n.Decode(&m); err != nil {
			return err
		}
		for k, v := range m {
			l.byLocale[k] = v
		}
		return nil
	default:
		return fmt.Errorf("localized: expected string or map, got kind %d", n.Kind)
	}
}

// Text 按 code 取文案，缺省回退到 en、再到任意非空项。
func (l LocalizedString) Text(code string) string {
	if s := l.byLocale[code]; s != "" {
		return s
	}
	if s := l.byLocale["en"]; s != "" {
		return s
	}
	for _, s := range l.byLocale {
		if s != "" {
			return s
		}
	}
	return ""
}

// AllValues 用于搜索，拼接所有非空语言文案。
func (l LocalizedString) AllValues() []string {
	var out []string
	seen := make(map[string]struct{})
	for _, s := range l.byLocale {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// HasText 是否至少配置了一种非空显示名（校验用）。
func (l LocalizedString) HasText() bool {
	return l.Text("en") != ""
}
