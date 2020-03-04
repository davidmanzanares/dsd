package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/davidmanzanares/dsd/provider"
)

type S3 struct {
	path   string
	region string
}

func parseURL(s string) (bucket string, key string, err error) {
	if !strings.HasPrefix(s, "s3://") {
		return "", "", fmt.Errorf("S3 service must begin with s3://, not with %s", s)
	}
	s = s[len("s3://"):]
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 1 {
		parts = append(parts, "/")
	}
	bucket = parts[0]
	key = parts[1]
	return bucket, key, err
}

func Create(service string) (provider.Provider, error) {
	bucket, _, err := parseURL(service)
	if err != nil {
		return nil, err
	}
	region, err := s3manager.GetBucketRegion(context.Background(), session.Must(session.NewSession()), bucket, "us-west-2")
	if err != nil {
		return nil, err
	}
	return &S3{path: service, region: region}, nil
}

func (s *S3) GetAsset(name string, writer io.Writer) error {
	bucket, key, err := parseURL(s.path + "/" + name)
	if err != nil {
		return err
	}

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(s.region)})

	downloader := s3manager.NewDownloader(sess)

	buff := &aws.WriteAtBuffer{}
	_, err = downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	writer.Write(buff.Bytes())

	return nil
}

func (s *S3) PushAsset(name string, reader io.Reader) error {
	return s.push("/assets/"+name, reader)
}

func (s *S3) push(name string, reader io.Reader) error {
	bucket, key, err := parseURL(s.path + name)
	if err != nil {
		return err
	}

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(s.region)})

	uploader := s3manager.NewUploader(sess)

	log.Println(bucket, " a  ", key)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3) PushVersion(v provider.Version) error {
	buff, err := v.Serialize()
	if err != nil {
		return err
	}
	return s.push("VERSION", bytes.NewReader(buff))
}
