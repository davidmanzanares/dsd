package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	log.SetFlags(log.Lshortfile)
	conf, _ := loadConfig()

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

	cmdDeploy := &cobra.Command{
		Use:   "deploy <target>",
		Short: "Deploys to <target>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target, ok := conf.Targets[args[0]]
			if !ok {
				fmt.Printf("Target \"%s\" doesn't exist\n", args[0])
			}
			fmt.Println("Deploying to", target)
			err := Deploy(*target)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	rootCmd.AddCommand(cmdDeploy)

	cmdDownload := &cobra.Command{
		Use:   "download <service>",
		Short: "Downloads the current deployment on <service>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := Download(args[0])
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	rootCmd.AddCommand(cmdDownload)

	cmdWatch := &cobra.Command{
		Use:   "watch <service>",
		Short: "Get <service> deployments, deploying the existing and new deployments",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Watch(args[0])
		},
	}
	rootCmd.AddCommand(cmdWatch)

	rootCmd.Execute()
}

func Deploy(target Target) error {
	p, err := s3.Create(target.Service)
	if err != nil {
		return err
	}

	/*gzipOutput, err := os.Create("out.tar.gzip")
	if err != nil {
		return err
	}*/

	uid := time.Now().UTC().Format(time.RFC3339) + " #" + hex.EncodeToString(uid())
	providerInput, gzipOutput := io.Pipe()
	gzipInput := gzip.NewWriter(gzipOutput)
	tarInput := tar.NewWriter(gzipInput)
	var pushError error
	var barrier sync.WaitGroup
	barrier.Add(1)
	go func() {
		pushError = p.PushAsset(uid+".tar.gz", providerInput)
		barrier.Done()
	}()
	numExecutables := 0
	for _, p := range target.Patterns {
		matches, err := filepath.Glob(p)
		if err != nil {
			log.Fatalln(err)
		}
		for _, path := range matches {
			f, err := os.Open(path)
			if err != nil {
				log.Println(err)
				continue
			}
			defer f.Close()
			fi, err := f.Stat()
			if err != nil {
				log.Println(err)
				continue
			}
			isExecutable := (fi.Mode() & 0100) != 0
			if isExecutable {
				numExecutables++
			}
			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				log.Println(err)
				continue
			}
			tarInput.WriteHeader(hdr)
			_, err = io.Copy(tarInput, f)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("copied", path)
		}
	}
	if numExecutables == 0 {
		fmt.Println("No executables were found")
		return nil
	}
	err = tarInput.Close()
	if err != nil {
		return err
	}
	err = gzipInput.Close()
	if err != nil {
		return err
	}
	gzipOutput.Close()
	if err != nil {
		return err
	}
	barrier.Wait()
	if pushError != nil {
		return pushError
	}

	err = p.PushVersion(provider.Version{Name: uid, Time: time.Now()})
	if err != nil {
		return err
	}
	fmt.Println("Deployed")
	return nil
}

func Watch(service string) {
	p, err := s3.Create(service)
	if err != nil {
		log.Fatal(err)
	}

	var currentVersion provider.Version
	var spawned *os.Process
	for {
		v, err := p.GetCurrentVersion()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		if v == currentVersion {
			time.Sleep(5 * time.Second)
			continue
		}
		exe, err := download(p, v)
		if err != nil {
			log.Println(err)
		} else {
			if spawned != nil {
				spawned.Signal(os.Interrupt)
				time.Sleep(time.Second)
				spawned.Kill()
			}
			time.Sleep(time.Second)
			fmt.Println("eee", exe)
			wd, err := os.Getwd()
			if err != nil {
				log.Println(err)
				continue
			}
			spawned, err = os.StartProcess(path.Join(wd, exe), []string{exe}, &os.ProcAttr{Dir: path.Dir(exe),
				Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
			if err != nil {
				log.Println(err)
				continue
			}
			currentVersion = v
		}
		time.Sleep(5 * time.Second)
	}
}

func Download(service string) error {
	p, err := s3.Create(service)
	if err != nil {
		return err
	}

	v, err := p.GetCurrentVersion()
	if err != nil {
		return err
	}
	_, err = download(p, v)
	return err
}

func download(p provider.Provider, v provider.Version) (string, error) {
	gzipInput, s3Output := io.Pipe()
	var barrier sync.WaitGroup
	barrier.Add(1)
	var err2 error
	go func() {
		err2 = p.GetAsset(v.Name+".tar.gz", s3Output)
		s3Output.Close()
		barrier.Done()
	}()

	gzipOutput, err := gzip.NewReader(gzipInput)
	if err != nil {
		return "", err
	}

	tarReader := tar.NewReader(gzipOutput)

	folder := "assets/" + v.Name + "/"
	err = os.MkdirAll(folder, 0770)
	if err != nil {
		return "", err
	}
	var executableFilepath string
	for {
		h, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		fmt.Println(h.Name, os.FileMode(h.Mode))
		path := folder + h.Name
		if h.Mode&100 != 0 && executableFilepath == "" {
			executableFilepath = path
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(h.Mode))
		if err != nil {
			log.Println(err)
			continue
		}
		defer f.Close()
		io.Copy(f, tarReader)
	}
	barrier.Wait()
	if err2 != nil {
		return "", err
	}
	return executableFilepath, nil
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

func saveConfig(conf Config) error {
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
