package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/davidmanzanares/dsd/provider"
)

func TestAsset(t *testing.T) {
	s, err := Create("s3://dsd-s3-test/" + base64.RawURLEncoding.EncodeToString(uid()))
	if err != nil {
		t.Fatal(err)
	}

	p := hex.EncodeToString(uid())

	err = s.PushAsset(p, strings.NewReader("holamundo"))
	if err != nil {
		t.Fatal(err)
	}

	var buff bytes.Buffer
	s.GetAsset(p, &buff)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVersion(t *testing.T) {
	s, err := Create("s3://dsd-s3-test/" + hex.EncodeToString(uid()))
	if err != nil {
		t.Fatal(err)
	}
	time := time.Now()
	name := hex.EncodeToString(uid())
	s.PushVersion(provider.Version{Name: name, Time: time})
	v, err := s.GetCurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != name {
		t.Errorf("Returned version name mismatch, expected %s, but got %s", name, v.Name)
	}
	if v.Name != name {
		t.Error("Returned version timestamp mismatch, expected ", t, ", but got", v.Time)
	}
}

func uid() []byte {
	buff := make([]byte, 6)
	rand.Read(buff)
	return buff
}
