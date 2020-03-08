package dsdl

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/davidmanzanares/dsd/types"
)

var testPatterns []string = []string{"test-asset*", "*/*", "*/*/*"}

func TestRunBasic(t *testing.T) {
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Run("s3://dsd-s3-test/tests", RunConf{})
	if err != nil {
		t.Fatal(err)
	}
	ev := r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppExit {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != Stopped {
		t.Fatal(ev)
	}
	checkFiles(v, t)
	checkExecution(t, v, 1)
}

func TestRunFailureRestart(t *testing.T) {
	var testPatterns []string = []string{"test-asset-failure-script", "*/*", "*/*/*"}
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Run("s3://dsd-s3-test/tests", RunConf{OnFailure: Restart})
	if err != nil {
		t.Fatal(err)
	}

	// First execution
	ev := r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppExit {
		t.Fatal(ev)
	}

	// Second execution
	ev = r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppExit {
		t.Fatal(ev)
	}
	r.Stop()
	execs := 2
	for {
		ev := r.WaitForEvent()
		if ev.Type == Stopped {
			break
		}
		if ev.Type == AppExit {
			execs++
		}
	}
	checkExecution(t, v, execs)
}

func TestRunHotReload(t *testing.T) {
	var testPatterns []string = []string{"test-asset-sleep-script", "*/*", "*/*/*"}
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Run("s3://dsd-s3-test/tests", RunConf{HotReload: true, Polling: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	ev := r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	time.Sleep(50 * time.Millisecond)
	checkExecution(t, v, 1)
	v, err = Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	time.Sleep(50 * time.Millisecond)
	checkExecution(t, v, 2)
	r.Stop()
	ev = r.WaitForEvent()
	if ev.Type != Stopped {
		t.Fatal(ev)
	}
}

func TestRunSuccessWait(t *testing.T) {
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Run("s3://dsd-s3-test/tests", RunConf{OnSuccess: Wait})
	if err != nil {
		t.Fatal(err)
	}
	ev := r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppExit {
		t.Fatal(ev)
	}
	checkFiles(v, t)
	checkExecution(t, v, 1)
	v, err = Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	r.commands <- "update"
	ev = r.WaitForEvent()
	if ev.Type != AppStarted {
		t.Fatal(ev)
	}
	ev = r.WaitForEvent()
	if ev.Type != AppExit {
		t.Fatal(ev)
	}
	checkFiles(v, t)
	checkExecution(t, v, 2)
	r.Stop()
	ev = r.WaitForEvent()
	if ev.Type != Stopped {
		t.Fatal(ev)
	}
}

func TestDefaultPolling(t *testing.T) {
	r, err := Run("s3://dsd-s3-test/tests", RunConf{OnSuccess: Wait})
	if err != nil {
		t.Fatal(err)
	}
	if r.conf.Polling != DefaultPolling {
		t.Fatal("Wrong polling time")
	}
	r.Stop()
}
func TestCustomPolling(t *testing.T) {
	r, err := Run("s3://dsd-s3-test/tests", RunConf{OnSuccess: Wait, Polling: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if r.conf.Polling != time.Minute {
		t.Fatal("Wrong polling time")
	}
	r.Stop()
}

func checkExecution(t *testing.T, v types.Version, executionTimes int) {
	d, err := ioutil.ReadFile("./assets/test-script-output")
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("called from %s#%d\n", file, no)
		}
		t.Fatal(err)
	}
	expected := strings.Repeat("I ran\n", executionTimes)
	if string(d) != expected {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("called from %s#%d\n", file, no)
		}
		fmt.Println(executionTimes)
		t.Fatal("test-script-output unexpected result:", string(d), string(expected), d, []byte(expected))
	}
}
