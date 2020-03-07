package dsdl

type Config struct {
	Targets map[string]*Target
}

type Target struct {
	Name     string `json:"-"`
	Service  string
	Patterns []string
}

func AddTarget(target Target) error {
	conf, _ := LoadConfig()
	conf.Targets[target.Name] = &target
	return SaveConfig(conf)
}
