package types

import (
	"testing"
	"time"
)

func TestVersionSerialization(t *testing.T) {
	v := Version{Name: "asd", Time: time.Now().Truncate(time.Second)}
	b, err := v.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	v2, err := DeserializeVersion(b)
	if err != nil {
		t.Fatal(err)
	}
	if v != v2 {
		t.Fatal(v2)
	}
}
