package dsdl

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/davidmanzanares/dsd/provider"
)

func TestMain(m *testing.M) {
	code := m.Run()
	deleteTestAssets()
	os.RemoveAll("./assets")
	os.Exit(code)
}

func TestDeployDownload(t *testing.T) {
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset*"}})
	if err != nil {
		t.Fatal(err)
	}
	err = Download("s3://dsd-s3-test/tests")
	if err != nil {
		t.Fatal(err)
	}
	checkFiles(v, t)
}

func checkFiles(v provider.Version, t *testing.T) {
	checkFile := func(filename string, expected string) {
		d, err := ioutil.ReadFile("./assets/" + v.Name + "/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		if string(d) != expected {
			t.Fatal("test-script-output unexpected result:", string(d), d, []byte(expected))
		}
	}
	checkFile("test-asset-1", "AssetA")
	checkFile("test-asset-2", "AssetB")
}
func checkExecution(v provider.Version, t *testing.T) {
	d, err := ioutil.ReadFile("./assets/" + v.Name + "/test-script-output")
	if err != nil {
		t.Fatal(err)
	}
	expected := "I ran\n"
	if string(d) != expected {
		t.Fatal("test-script-output unexpected result:", string(d), string(d), d, []byte(expected))
	}
}

func TestDeployWatch(t *testing.T) {
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset*"}})
	if err != nil {
		t.Fatal(err)
	}
	w := Watch("s3://dsd-s3-test/tests", nil)
	defer w.Stop()
	w.Poll()
	checkFiles(v, t)
	time.Sleep(100 * time.Millisecond)
	checkExecution(v, t)
	v, err = Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset*"}})
	if err != nil {
		t.Fatal(err)
	}
	w.Poll()
	w.Poll()
	checkFiles(v, t)
	time.Sleep(100 * time.Millisecond)
	checkExecution(v, t)
}

func TestAddTarget(t *testing.T) {
	// TODO
}

func TestDeployFailureNoExecutable(t *testing.T) {
	service := "s3://dsd-s3-test/tests"
	_, err := Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset-1"}})
	if err == nil {
		t.Fatal("Deploy should fail when there is no executable")
	}
}

func TestDeployServiceFailure(t *testing.T) {
	service := "s3://dsd-s3-test-invalid/tests"
	_, err := Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset*"}})
	if err == nil {
		t.Fatal("Deploy should fail when the service URL is invalid")
	}
}

func createTestAssets() {
	err := ioutil.WriteFile("test-asset-1", []byte("AssetA"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-2", []byte("AssetB"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-script", []byte(
		`#!/bin/sh
		
		echo "I ran" >> test-script-output`), 0770)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteTestAssets() {
	os.Remove("test-asset-1")
	os.Remove("test-asset-2")
	os.Remove("test-asset-script")
	os.Remove("test-script-output")
}
