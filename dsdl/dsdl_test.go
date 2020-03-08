package dsdl

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/davidmanzanares/dsd/types"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)
	code := m.Run()
	deleteTestAssets()
	os.Exit(code)
}

func createTestAssets() {
	err := os.MkdirAll("test-asset-basic-folder", 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("test-asset-basic-folder/folder2", 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-basic-1", []byte("AssetA"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-basic-2", []byte("AssetB"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-basic-folder/test-asset-3", []byte("AssetC"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-basic-folder/folder2/test-asset-4", []byte("AssetD"), 0660)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-basic-script", []byte(
		`#!/bin/sh
		echo "I ran" >> ../test-script-output`), 0770)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-sleep-script", []byte(
		`#!/bin/sh
		echo "I ran" >> ../test-script-output
		sleep 30`), 0770)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test-asset-failure-script", []byte(
		`#!/bin/sh
		echo "I ran" >> ../test-script-output
		exit 123`), 0770)
	if err != nil {
		log.Fatal(err)
	}
}

func checkFiles(v types.Version, t *testing.T) {
	checkFile := func(filename string, expected string) {
		d, err := ioutil.ReadFile("./assets/" + v.Name + "/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		if string(d) != expected {
			t.Fatal("test-script-output unexpected result:", string(d), d, []byte(expected))
		}
	}
	checkFile("test-asset-basic-1", "AssetA")
	checkFile("test-asset-basic-2", "AssetB")
	checkFile("test-asset-basic-folder/test-asset-3", "AssetC")
	checkFile("test-asset-basic-folder/folder2/test-asset-4", "AssetD")
}

func deleteTestAssets() {
	os.RemoveAll("test-asset-basic-folder")
	os.RemoveAll("assets")
	os.Remove("test-asset-basic-1")
	os.Remove("test-asset-basic-2")
	os.Remove("test-asset-basic-script")
	os.Remove("test-asset-sleep-script")
	os.Remove("test-asset-failure-script")
}
