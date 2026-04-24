package config

// Tool 为单条可装工具。name、desc 支持 string（视为 en）或多语言表。
type Tool struct {
	ID        string          `yaml:"id"`
	Name      LocalizedString `yaml:"name"`
	Desc      LocalizedString `yaml:"desc"`
	Check     string          `yaml:"check"`
	Install   InstallCommands `yaml:"install"`
	Installed bool            `yaml:"-"`
	Supported bool            `yaml:"-"`
}

type InstallCommands struct {
	Windows string `yaml:"windows"`
	Darwin  string `yaml:"darwin"`
	Debian  string `yaml:"debian"`
}
