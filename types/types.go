package types

import (
	"encoding/json"
	"io"
	"time"
)

// Provider implementations contains the necessary functions to
// deploy and download assets from it
type Provider interface {
	GetAsset(name string, writer io.Writer) error
	PushAsset(name string, reader io.Reader) error

	PushVersion(v Version) error
	GetCurrentVersion() (Version, error)
}

// Version is composed of a unique name (identifier) and a timestamp
type Version struct {
	Name string
	Time time.Time
}

// Serialize marshals v
func (v Version) Serialize() ([]byte, error) {
	return json.Marshal(v)
}

// DeserializeVersion unmarshal the version stored in b
func DeserializeVersion(b []byte) (v Version, err error) {
	err = json.Unmarshal(b, &v)
	return v, err
}
