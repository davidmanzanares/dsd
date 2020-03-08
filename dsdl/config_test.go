package dsdl

import (
	"os"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	w, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := "./test-config"
	os.Mkdir(dir, 0777)
	defer os.RemoveAll(dir)
	os.Chdir(dir)
	defer os.Chdir(w)

	conf, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error, no config file")
	}
	if !reflect.DeepEqual(conf, Config{Targets: make(map[string]*Target)}) {
		t.Fatal(conf)
	}

	target := Target{Name: "myname", Patterns: []string{"asd", "wadus"}, Service: "myservice"}
	err = AddTarget(target)
	if err != nil {
		t.Fatal(err)
	}

	conf, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	expected := make(map[string]*Target)
	expected["myname"] = &target
	if !reflect.DeepEqual(conf, Config{Targets: expected}) {
		t.Fatal(conf)
	}
}
