package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestUploadDownload(t *testing.T) {
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

func uid() []byte {
	buff := make([]byte, 6)
	rand.Read(buff)
	return buff
}
