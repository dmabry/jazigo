package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// hasPrintf is an interface for loggers that support Printf
type hasPrintf interface {
	Printf(string, ...interface{})
}

var s3SvcTable = map[string]*s3.Client{} // region => client
var s3logger hasPrintf
var s3region string // default region

func s3client(region string) *s3.Client {
	if region == "" {
		region = s3region // fallback to default region
		if region == "" {
			s3log("s3client: could not find region")
			return nil
		}
	}

	svc, ok := s3SvcTable[region]
	if !ok {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region))
		if err != nil {
			s3log("s3client: could not load config for region %s: %v", region, err)
			return nil
		}

		svc = s3.NewFromConfig(cfg)
		s3SvcTable[region] = svc
		s3log("s3client: client created: region=[%s]", region)
	}

	return svc
}

func s3init(logger hasPrintf, region string) {
	if s3logger != nil {
		panic("s3 store reinitialization")
	}
	if logger == nil {
		panic("s3 store nil logger")
	}
	s3region = region
	s3logger = logger
	s3log("initialized: default region=[%s]", s3region)
}

func s3log(format string, v ...interface{}) {
	if s3logger == nil {
		log.Printf("s3 store (uninitialized): "+format, v...)
		return
	}
	s3logger.Printf("s3 store: "+format, v...)
}

// S3Path checks if path is an aws s3 path.
func S3Path(path string) bool {
	return s3path(path)
}

func s3path(path string) bool {
	s3match := strings.HasPrefix(path, "arn:aws:s3:")
	if s3match {
		s3log("s3path: [%s]", path)
	}
	return s3match
}

//	Input: "arn:aws:s3:region::bucket/folder/file.xxx"
//
// Output: "region", "bucket", "folder/file.xxx"
func s3parse(path string) (string, string, string) {
	s := strings.Split(path, ":")
	if len(s) < 6 {
		return "", "", ""
	}
	region := s[3]
	file := s[5]
	slash := strings.IndexByte(file, '/')
	if slash < 1 {
		return "", "", ""
	}
	bucket := file[:slash]
	key := file[slash+1:]
	return region, bucket, key
}

func s3fileExists(path string) bool {

	region, bucket, key := s3parse(path)

	svc := s3client(region)
	if svc == nil {
		s3log("s3fileExists: missing s3 client: ugh")
		return false // ugh
	}

	input := &s3.HeadObjectInput{
		Bucket: &bucket, // Required
		Key:    &key,    // Required
	}
	if _, err := svc.HeadObject(context.TODO(), input); err == nil {
		//s3log("s3fileExists: FOUND [%s]", path)
		return true // found
	}

	return false
}

func s3fileput(path string, buf []byte, contentType string) error {

	region, bucket, key := s3parse(path)

	svc := s3client(region)
	if svc == nil {
		return fmt.Errorf("s3fileput: missing s3 client")
	}

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(buf),
	}

	switch contentType {
	case "": // none
	case "detect": // detect
		contentTypeDetected := http.DetectContentType(buf)
		input.ContentType = &contentTypeDetected
	default: // use literal
		input.ContentType = &contentType
	}

	_, err := svc.PutObject(context.TODO(), input)

	//s3log("s3fileput: [%s] upload: error: %v", path, err)

	return err
}

func s3fileRemove(path string) error {

	region, bucket, key := s3parse(path)

	svc := s3client(region)
	if svc == nil {
		return fmt.Errorf("s3fileRemove: missing s3 client")
	}

	input := &s3.DeleteObjectInput{
		Bucket: &bucket, // Required
		Key:    &key,    // Required
	}
	_, err := svc.DeleteObject(context.TODO(), input)

	//s3log("s3fileRemove: [%s] delete: error: %v", path, err)

	return err
}

func s3fileRename(p1, p2 string) error {

	region, bucket1, key1 := s3parse(p1)

	svc := s3client(region)
	if svc == nil {
		return fmt.Errorf("s3fileRename: missing s3 client")
	}

	_, bucket2, key2 := s3parse(p2)

	copySource := fmt.Sprintf("%s/%s", bucket1, key1) // Required
	copySourcePtr := &copySource

	input := &s3.CopyObjectInput{
		Bucket:     &bucket2, // Required
		CopySource: copySourcePtr,
		Key:        &key2, // Required
	}
	_, copyErr := svc.CopyObject(context.TODO(), input)
	if copyErr != nil {
		return copyErr
	}

	if removeErr := s3fileRemove(p1); removeErr != nil {
		// could not remove old file
		s3fileRemove(p2) // remove new file (clean up)
		return removeErr
	}

	return nil
}

func s3fileReader(path string) (io.ReadCloser, error) {

	region, bucket, key := s3parse(path)

	svc := s3client(region)
	if svc == nil {
		return nil, fmt.Errorf("s3fileRead: missing s3 client")
	}

	input := &s3.GetObjectInput{
		Bucket: &bucket, // Required
		Key:    &key,    // Required
	}

	resp, err := svc.GetObject(context.TODO(), input)

	return resp.Body, err
}

// S3URL builds the URL for an S3 bucket.
// The path is an ARN: "arn:aws:s3:region::bucket/folder/file.xxx"
func S3URL(path string) string {
	region, bucket, key := s3parse(path)

	if region == "" {
		region = s3region // fallback to default region
	}
	if region == "" {
		s3log("S3URL: could not find region: [%s]", path)
		return ""
	}

	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", region, bucket, key)
}
