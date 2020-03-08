package dsdl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/davidmanzanares/dsd/provider"
	"github.com/davidmanzanares/dsd/provider/s3"
)

func getProviderFromService(service string) (provider.Provider, error) {
	if strings.HasPrefix(service, "s3:") {
		return s3.Create(service)
	}
	return nil, errors.New(fmt.Sprint("Unkown service:", service))
}
