package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidmanzanares/dsd/provider"
	"github.com/davidmanzanares/dsd/provider/s3"
)

type Config struct {
	Targets map[string]*Target
}

type Target struct {
	Name     string `json:"-"`
	Service  string
	Patterns []string
}

func (t Target) String() string {
	var patterns []string
	for _, p := range t.Patterns {
		patterns = append(patterns, `"`+p+`"`)
	}
	return fmt.Sprintf("\"%s\" (%s) {%s}", t.Name, t.Service, strings.Join(patterns, ", "))
}

func main() {
	conf, err := loadConfig()

	cmdAdd := &cobra.Command{
		Use:   "add <target> <service> <pattern1> [patterns2]...",
		Short: "Add a new target to deploy",
		Long:  `Adds a new target to deploy, a target is composed by its name, its service URL and a list of glob patterns.`,
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			conf, err := loadConfig()
			target := Target{Name: args[0], Service: args[1], Patterns: args[2:]}
			conf.Targets[target.Name] = &target
			err = saveConfig(conf)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("Target %s added\n", target)
		},
	}
	rootCmd := &cobra.Command{Use: "dsd <command>"}
	rootCmd.AddCommand(cmdAdd)

	if err == nil {
		for _, target := range conf.Targets {
			cmdDeploy := &cobra.Command{
				Use:   target.Name,
				Short: fmt.Sprint("Deploys to \"%s\"", target.Name),
				Long:  fmt.Sprint(`Deploys to \"%s\" by using "s3:"`, target.Name),
				Args:  cobra.ExactArgs(0),
				Run: func(cmd *cobra.Command, args []string) {
					fmt.Println("Deploying to", target.Name)
					matches, err := filepath.Glob("./provider/*")
					fmt.Println(matches, err)
				},
			}
			rootCmd.AddCommand(cmdDeploy)

		}
	}
	fmt.Println(conf)
	rootCmd.Execute()
}

func Publish(target Target) {
	// Compress assets under path
	// Upload compressed assets
	// Push version
}

func Watch(target Target) {
	// ListVersions
	// Get latest version
	// Get asset
	// decompress
	// Stop
	// Play
}

func getProviderFromService(service string) (provider.Provider, error) {
	if strings.HasPrefix(service, "s3:") {
		return s3.Create(service)
	}
	return nil, errors.New(fmt.Sprint("Unkown service:", service))
}

func loadConfig() (Config, error) {
	buffer, err := ioutil.ReadFile(".dsd.json")
	if err != nil {
		return Config{}, err
	}
	var conf Config
	conf.Targets = make(map[string]*Target)
	err = json.Unmarshal(buffer, &conf)
	for k, _ := range conf.Targets {
		conf.Targets[k].Name = k
	}
	fmt.Println(22, conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func saveConfig(conf Config) error {
	buffer, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}
	ioutil.WriteFile(".dsd.json", buffer, 0660)
	return nil
}
