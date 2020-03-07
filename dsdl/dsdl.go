package dsdl

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

func Deploy(target Target) (provider.Version, error) {
	p, err := s3.Create(target.Service)
	if err != nil {
		return provider.Version{}, err
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

	folders := make(map[string]bool)

	numExecutables := 0
	for _, p := range target.Patterns {
		matches, err := filepath.Glob(p)
		if err != nil {
			return provider.Version{}, err
		}
		for _, filepath := range matches {
			func() {
				dir := path.Dir(filepath)
				for dir != "." {
					if !folders[dir] {
						folders[dir] = true
						fi, err := os.Stat(dir)
						if err != nil {
							hdr, err := tar.FileInfoHeader(fi, "")
							if err != nil {
								log.Println(err)
							} else {
								defer tarInput.WriteHeader(hdr)
							}
						} else {
							log.Println(err)
						}
					}
					dir = path.Dir(filepath)
				}
			}()
			f, err := os.Open(filepath)
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
			hdr.Name = filepath
			tarInput.WriteHeader(hdr)
			_, err = io.Copy(tarInput, f)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
	if numExecutables == 0 {
		return provider.Version{}, errors.New("No executables")
	}
	err = tarInput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	err = gzipInput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	gzipOutput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	barrier.Wait()
	if pushError != nil {
		return provider.Version{}, pushError
	}

	v := provider.Version{Name: uid, Time: time.Now()}
	err = p.PushVersion(v)
	if err != nil {
		return provider.Version{}, err
	}
	return v, nil
}

type Watcher struct {
	shouldContinue chan bool
	stopped        chan bool
	ExitCode       chan int
}

func (w *Watcher) Poll() {
	w.shouldContinue <- true
}
func (w *Watcher) Stop() {
	w.shouldContinue <- false
	<-w.stopped
}

func Watch(service string, args []string) *Watcher {
	p, err := s3.Create(service)
	if err != nil {
		log.Fatal(err)
	}

	w := &Watcher{shouldContinue: make(chan bool), stopped: make(chan bool), ExitCode: make(chan int)}
	go func() {
		var currentVersion provider.Version
		var spawned *os.Process
		work := func() {
			v, err := p.GetCurrentVersion()
			if err != nil {
				log.Println(err)
				return
			}
			if v == currentVersion {
				return
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
				wd, err := os.Getwd()
				if err != nil {
					log.Println(err)
					return
				}
				spawned, err = os.StartProcess(path.Join(wd, exe), append([]string{exe}, args...), &os.ProcAttr{Dir: path.Dir(exe),
					Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
				if err != nil {
					log.Println(err)
					return
				}
				go func(ch chan int) {
					state, _ := spawned.Wait()
					ch <- state.ExitCode()
				}(w.ExitCode)

				currentVersion = v
			}
		}
		for {
			select {
			case <-time.After(5 * time.Second):
				work()
			case shouldContinue := <-w.shouldContinue:
				if shouldContinue {
					work()
					continue
				}
				if spawned != nil {
					spawned.Signal(os.Interrupt)
					time.Sleep(time.Second)
					spawned.Kill()
				}
				w.stopped <- true
				return
			}

		}
	}()
	w.shouldContinue <- true
	return w
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
		filepath := folder + h.Name

		if h.FileInfo().IsDir() {
			os.Mkdir(filepath, h.FileInfo().Mode().Perm())
		}

		if h.Mode&0100 != 0 && executableFilepath == "" {
			executableFilepath = filepath
		}

		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(h.Mode))
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
