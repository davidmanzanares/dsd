package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/davidmanzanares/dsd/types"
)

type S3 struct {
	path   string
	region string
}

func Create(service string) (types.Provider, error) {
	bucket, _, err := parseURL(service)
	if err != nil {
		return nil, err
	}
	region, err := s3manager.GetBucketRegion(context.Background(), session.Must(session.NewSession()), bucket, "eu-west-1")
	if err != nil {
		return nil, err
	}
	return &S3{path: service, region: region}, nil
}

func (s *S3) GetAsset(name string, writer io.Writer) error {
	return s.get("/assets/"+name, writer)
}

func (s *S3) PushAsset(name string, reader io.Reader) error {
	return s.push("/assets/"+name, reader)
}

func (s *S3) GetCurrentVersion() (types.Version, error) {
	buffer := bytes.NewBuffer(nil)
	s.get("/VERSION", buffer)
	return types.UnserializeVersion(buffer.Bytes())
}

func (s *S3) PushVersion(v types.Version) error {
	buff, err := v.Serialize()
	if err != nil {
		return err
	}
	return s.push("/VERSION", bytes.NewReader(buff))
}

func (s *S3) get(path string, writer io.Writer) error {
	bucket, key, err := parseURL(s.path + path)
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

func (s *S3) push(name string, reader io.Reader) error {
	bucket, key, err := parseURL(s.path + name)
	if err != nil {
		return err
	}

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(s.region)})

	uploader := s3manager.NewUploader(sess)

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

var invalidURL error = errors.New("S3 service must begin with s3://")

func parseURL(s string) (bucket string, key string, err error) {
	if !strings.HasPrefix(s, "s3://") {
		return "", "", invalidURL
	}
	s = s[len("s3://"):]
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 1 {
		parts = append(parts, "")
	}
	bucket = parts[0]
	key = parts[1]
	return bucket, key, err
}
