package dsdl

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Config is a set of targets
type Config struct {
	Targets map[string]*Target
}

// Target is a combination of an alias name to deploy,
// a provider service,
// and a list of glob patterns
type Target struct {
	Name     string `json:"-"`
	Service  string
	Patterns []string
}

// AddTarget loads the config from the default path, adds the new target, and saves the new config file
func AddTarget(target Target) error {
	conf, _ := LoadConfig()
	conf.Targets[target.Name] = &target
	return SaveConfig(conf)
}

func (t Target) String() string {
	var patterns []string
	for _, p := range t.Patterns {
		patterns = append(patterns, `"`+p+`"`)
	}
	return fmt.Sprintf("\"%s\" (%s) {%s}", t.Name, t.Service, strings.Join(patterns, ", "))
}

// LoadConfig loads the config from the default path
func LoadConfig() (Config, error) {
	var conf Config
	conf.Targets = make(map[string]*Target)
	buffer, err := ioutil.ReadFile(".dsd.json")
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(buffer, &conf)
	for k := range conf.Targets {
		conf.Targets[k].Name = k
	}
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// SaveConfig saves conf on the default path
func SaveConfig(conf Config) error {
	buffer, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}
	ioutil.WriteFile(".dsd.json", buffer, 0660)
	return nil
}

func uid() []byte {
	buff := make([]byte, 8)
	rand.Read(buff)
	return buff
}
