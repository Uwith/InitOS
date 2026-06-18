package prefs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	appDirName  = "initos"
	localeFName = "locale"
)

// ConfigDir 返回 $XDG_CONFIG_HOME/initos 或 用户配置目录下的 initos（跨平台与 Go 1.13+ UserConfigDir 一致）。
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName), nil
}

// ReadLocale 读取已保存语言码（单文件单行），不存在则返回 ("", false)。
func ReadLocale() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, localeFName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	code := strings.TrimSpace(string(data))
	if code == "" {
		return "", nil
	}
	return code, nil
}

// WriteLocale 持久化语言选择。
func WriteLocale(code string) error {
	if code == "" {
		return nil
	}
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, localeFName), []byte(code+"\n"), 0o600)
}
