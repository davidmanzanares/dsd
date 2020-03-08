package dsdl

import "testing"

func TestDeployFailureNoExecutable(t *testing.T) {
	service := "s3://dsd-s3-test/tests"
	_, err := Deploy(Target{Name: "test", Service: service, Patterns: []string{"test-asset-1"}})
	if err == nil {
		t.Fatal("Deploy should fail when there is no executable")
	}
}
