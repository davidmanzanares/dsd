package dsdl

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

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

func (t Target) String() string {
	var patterns []string
	for _, p := range t.Patterns {
		patterns = append(patterns, `"`+p+`"`)
	}
	return fmt.Sprintf("\"%s\" (%s) {%s}", t.Name, t.Service, strings.Join(patterns, ", "))
}

func LoadConfig() (Config, error) {
	var conf Config
	conf.Targets = make(map[string]*Target)
	buffer, err := ioutil.ReadFile(".dsd.json")
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(buffer, &conf)
	for k, _ := range conf.Targets {
		conf.Targets[k].Name = k
	}
	if err != nil {
		return conf, err
	}
	return conf, nil
}

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
