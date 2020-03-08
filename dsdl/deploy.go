package dsdl

import (
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/davidmanzanares/dsd/provider"
	"github.com/davidmanzanares/dsd/provider/s3"
)

func Deploy(target Target) (provider.Version, error) {
	p, err := s3.Create(target.Service)
	if err != nil {
		return provider.Version{}, err
	}

	/*gzipOutput, err := os.Create("out.tar.gzip")
	if err != nil {
		return err
	}*/

	uid := time.Now().UTC().Format(time.RFC3339) + " #" + hex.EncodeToString(uid())
	providerInput, gzipOutput := io.Pipe()
	gzipInput := gzip.NewWriter(gzipOutput)
	tarInput := tar.NewWriter(gzipInput)
	var pushError error
	var barrier sync.WaitGroup
	barrier.Add(1)
	go func() {
		pushError = p.PushAsset(uid+".tar.gz", providerInput)
		barrier.Done()
	}()

	folders := make(map[string]bool)

	numExecutables := 0
	for _, p := range target.Patterns {
		matches, err := filepath.Glob(p)
		if err != nil {
			return provider.Version{}, err
		}
		for _, filepath := range matches {
			func() {
				dir := path.Dir("./" + filepath)

				for i := 0; dir != "."; i++ {
					if i == 1000 {
						panic(i)
					}
					if !folders[dir] {
						folders[dir] = true
						fi, err := os.Stat(dir)
						if err == nil {
							hdr, err := tar.FileInfoHeader(fi, "")
							if err != nil {
								log.Println(err)
							} else {
								hdr.Name = dir
								defer tarInput.WriteHeader(hdr)
							}
						} else {
							log.Println(err)
						}
					}
					dir = path.Dir(dir)
				}
			}()
			f, err := os.Open(filepath)
			if err != nil {
				log.Println(err)
				continue
			}
			defer f.Close()
			fi, err := f.Stat()
			if err != nil {
				log.Println(err)
				continue
			}
			if fi.IsDir() {
				continue
			}

			isExecutable := (fi.Mode() & 0100) != 0
			if isExecutable {
				numExecutables++
			}
			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				log.Println(err)
				continue
			}
			hdr.Name = filepath
			tarInput.WriteHeader(hdr)
			_, err = io.Copy(tarInput, f)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
	if numExecutables == 0 {
		return provider.Version{}, errors.New("No executables")
	}
	err = tarInput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	err = gzipInput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	gzipOutput.Close()
	if err != nil {
		return provider.Version{}, err
	}
	barrier.Wait()
	if pushError != nil {
		return provider.Version{}, pushError
	}

	v := provider.Version{Name: uid, Time: time.Now()}
	err = p.PushVersion(v)
	if err != nil {
		return provider.Version{}, err
	}
	return v, nil
}
