package dsdl

import "testing"

func TestDeployServiceFailure(t *testing.T) {
	service := "s3://dsd-s3-test-invalid/tests"
	_, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err == nil {
		t.Fatal("Deploy should fail when the service URL is invalid")
	}
}
