package dsdl

import "testing"

func TestGetProviderFromService(t *testing.T) {
	_, err := getProviderFromService("s3://dsd-s3-test/tests")
	if err != nil {
		t.Fatal(err)
	}
}
func TestGetProviderFromServiceInvalid(t *testing.T) {
	_, err := getProviderFromService("invalid://dsd-s3-test/tests")
	if err == nil {
		t.Fatal("Expected error")
	}
}
