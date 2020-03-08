package dsdl

import "testing"

func TestDeployDownload(t *testing.T) {
	createTestAssets()
	defer deleteTestAssets()

	service := "s3://dsd-s3-test/tests"
	v, err := Deploy(Target{Name: "test", Service: service, Patterns: testPatterns})
	if err != nil {
		t.Fatal(err)
	}
	err = Download("s3://dsd-s3-test/tests")
	if err != nil {
		t.Fatal(err)
	}
	checkFiles(v, t)
}
