package dsdl

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"sync"

	"github.com/davidmanzanares/dsd/types"
)

// Download the assets deployed on service
func Download(service string) error {
	p, err := getProviderFromService(service)
	if err != nil {
		return err
	}

	v, err := p.GetCurrentVersion()
	if err != nil {
		return err
	}
	_, err = download(p, v)
	return err
}

func download(p types.Provider, v types.Version) (string, error) {
	gzipInput, s3Output := io.Pipe()
	var barrier sync.WaitGroup
	barrier.Add(1)
	var err2 error
	go func() {
		err2 = p.GetAsset(v.Name+".tar.gz", s3Output)
		s3Output.Close()
		barrier.Done()
	}()

	gzipOutput, err := gzip.NewReader(gzipInput)
	if err != nil {
		return "", err
	}

	tarReader := tar.NewReader(gzipOutput)

	folder := "assets/" + v.Name + "/"
	err = os.MkdirAll(folder, 0770)
	if err != nil {
		return "", err
	}
	var executableFilepath string
	for {
		h, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		filepath := folder + h.Name

		if h.FileInfo().IsDir() {
			os.Mkdir(filepath, h.FileInfo().Mode().Perm())
			continue
		}

		if h.Mode&0100 != 0 && executableFilepath == "" {
			executableFilepath = filepath
		}

		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(h.Mode))
		if err != nil {
			log.Println(err)
			continue
		}
		defer f.Close()
		io.Copy(f, tarReader)
	}
	barrier.Wait()
	if err2 != nil {
		return "", err
	}
	return executableFilepath, nil
}
