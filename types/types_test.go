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
	if v.Name != v2.Name {
		t.Fatal(v.Name, "!=", v2.Name)
	}
	if !v.Time.Equal(v2.Time) {
		t.Fatal(v.Time, "!=", v2.Time)
	}
}
