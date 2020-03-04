package provider

import (
	"encoding/json"
	"io"
	"time"
)

type Provider interface {
	GetAsset(name string, writer io.Writer) error
	PushAsset(name string, reader io.Reader) error

	PushVersion(v Version) error
	GetCurrentVersion() (Version, error)
}

type Version struct {
	Name string
	Time time.Time
}

func (v Version) Serialize() ([]byte, error) {
	return json.Marshal(v)
}

func UnserializeVersion(b []byte) (v Version, err error) {
	err = json.Unmarshal(b, &v)
	return v, err
}
