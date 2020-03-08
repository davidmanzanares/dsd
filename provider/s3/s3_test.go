package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/davidmanzanares/dsd/types"
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
	s.PushVersion(types.Version{Name: name, Time: time})
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

func TestAccessError(t *testing.T) {
	_, err := Create("s3://dsd-s3-test-INVALID-BUCKET" + hex.EncodeToString(uid()))
	if err == nil {
		t.Fatal("No Access error")
	}
}

func TestParseURL1(t *testing.T) {
	testParseURL("s3://bucket", "bucket", "", nil, t)
	testParseURL("s3://bucket/", "bucket", "", nil, t)
}
func TestParseURL2(t *testing.T) {
	testParseURL("s3://bucket/folder1", "bucket", "folder1", nil, t)
}
func TestParseURL3(t *testing.T) {
	testParseURL("s3://bucket/folder1/folder2", "bucket", "folder1/folder2", nil, t)
}
func TestParseURLerror(t *testing.T) {
	testParseURL("s3", "", "", errInvalidURL, t)
}

func testParseURL(url, expectedBucket, expectedKey string, expectedError error, t *testing.T) {
	b, k, e := parseURL(url)
	if e != expectedError {
		t.Fatal(e)
	}
	if b != expectedBucket {
		t.Fatal(b)
	}
	if k != expectedKey {
		t.Fatal(k)
	}
}

func uid() []byte {
	buff := make([]byte, 6)
	rand.Read(buff)
	return buff
}
