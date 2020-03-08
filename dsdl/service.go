package dsdl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/davidmanzanares/dsd/provider/s3"
	"github.com/davidmanzanares/dsd/types"
)

func getProviderFromService(service string) (types.Provider, error) {
	if strings.HasPrefix(service, "s3:") {
		return s3.Create(service)
	}
	return nil, errors.New(fmt.Sprint("Unkown service:", service))
}
