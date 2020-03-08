package dsdl

import (
	"fmt"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/davidmanzanares/dsd/provider"
	"github.com/davidmanzanares/dsd/provider/s3"
)

type RunConf struct {
	Args      []string
	HotReload bool
	OnSuccess RunReaction
	OnFailure RunReaction
}

type RunReaction int

const (
	Exit RunReaction = iota
	Wait
	Restart
)

type Runner struct {
	events   chan RunEvent
	commands chan string

	conf           RunConf
	provider       provider.Provider
	currentVersion provider.Version
	appExe         string
	spawned        *os.Process
	exit           chan exitType
}
type exitType struct {
	code int
	v    provider.Version
}

type RunEventType int

const (
	AppStarted RunEventType = iota
	AppExit
	Stopped
)

type RunEvent struct {
	Type     RunEventType
	Version  provider.Version
	ExitCode int
}

func (e RunEvent) String() string {
	if e.Type == AppStarted {
		return fmt.Sprintf("AppStarted{v: %s}", e.Version)
	} else if e.Type == AppExit {
		return fmt.Sprintf("AppExit{Version: %s, ExitCode: %d}", e.Version, e.ExitCode)
	} else if e.Type == Stopped {
		return "Stopped"
	} else {
		panic(e)
	}
}

func Run(service string, conf RunConf) (*Runner, error) {
	p, err := s3.Create(service)
	if err != nil {
		return nil, err
	}

	r := &Runner{events: make(chan RunEvent, 10), commands: make(chan string, 10), provider: p, conf: conf}
	go r.manager()
	r.commands <- "update"
	return r, nil
}

// WaitForEvent waits for the generation of the next RunEvent
func (r *Runner) WaitForEvent() RunEvent {
	ev, ok := <-r.events
	if !ok {
		return RunEvent{Type: Stopped}
	}
	return ev
}

// Stop stops the current runner, interrupting/killing the application
func (r *Runner) Stop() {
	r.commands <- "stop"
}

func (r *Runner) manager() {
	defer close(r.events)
	for {
		select {
		case exit := <-r.exit:
			r.events <- RunEvent{Type: AppExit, Version: exit.v, ExitCode: exit.code}
			r.spawned = nil
			if exit.code == 0 {
				if r.conf.OnSuccess == Restart {
					r.run()
				} else if r.conf.OnSuccess == Exit {
					return
				} else {
					r.exit = nil
				}
			} else {
				if r.conf.OnFailure == Restart {
					r.run()
				} else if r.conf.OnFailure == Exit {
					return
				} else {
					r.exit = nil
				}
			}
		case <-time.After(5 * time.Second):
			if r.conf.HotReload || r.spawned == nil {
				r.update()
			}
		case cmd := <-r.commands:
			if cmd == "update" {
				r.update()
				continue
			} else if cmd == "stop" {
				r.kill()
				return
			} else {
				panic(fmt.Sprint("Unkown command:", cmd))
			}
		}
	}
}
func (r *Runner) kill() {
	if r.spawned != nil {
		// TODO call Interrupt first
		//r.spawned.Signal(os.Interrupt)
		//time.Sleep(time.Second)

		pgid, err := syscall.Getpgid(r.spawned.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15) // note the minus sign
		}

		r.spawned = nil
		for range r.exit {
		}
	}
}

func (r *Runner) update() {
	v, err := r.provider.GetCurrentVersion()
	if err != nil {
		log.Println(err)
		return
	}
	if v == r.currentVersion {
		return
	}
	exe, err := download(r.provider, v)
	if err != nil {
		log.Println(err)
		return
	}
	r.appExe = exe
	r.currentVersion = v
	r.run()
}

func (r *Runner) run() {
	r.kill()
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	// TODO windows should kill the process tree (grand children too)
	r.spawned, err = os.StartProcess(path.Join(wd, r.appExe), append([]string{r.appExe}, r.conf.Args...),
		&os.ProcAttr{
			Dir:   path.Dir(r.appExe),
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Sys:   &syscall.SysProcAttr{Setpgid: true}})
	if err != nil {
		log.Println(err)
		return
	}
	r.events <- RunEvent{Type: AppStarted, Version: r.currentVersion}
	r.exit = make(chan exitType)
	go func(spawned *os.Process, v provider.Version, exitCh chan exitType) {
		state, _ := spawned.Wait()
		exitCh <- exitType{code: state.ExitCode(), v: v}
		close(exitCh)
	}(r.spawned, r.currentVersion, r.exit)
}
