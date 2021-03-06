// +build integration

package s3manager

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/awstesting/integration"
	s3integ "github.com/IBM/ibm-cos-sdk-go/awstesting/integration/customizations/s3"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
)

var bucketName *string

func TestMain(m *testing.M) {
	svc := s3.New(integration.Session)
	bucketName = aws.String(s3integ.GenerateBucketName())
	if err := s3integ.SetupTest(svc, *bucketName); err != nil {
		panic(err)
	}

	var result int
	defer func() {
		if err := s3integ.CleanupTest(svc, *bucketName); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "S3 integrationt tests paniced,", r)
			result = 1
		}
		os.Exit(result)
	}()

	result = m.Run()
}

type dlwriter struct {
	buf []byte
}

func newDLWriter(size int) *dlwriter {
	return &dlwriter{buf: make([]byte, size)}
}

func (d dlwriter) WriteAt(p []byte, pos int64) (n int, err error) {
	if pos > int64(len(d.buf)) {
		return 0, io.EOF
	}

	written := 0
	for i, b := range p {
		if i >= len(d.buf) {
			break
		}
		d.buf[pos+int64(i)] = b
		written++
	}
	return written, nil
}

func validate(t *testing.T, key string, md5value string) {
	mgr := s3manager.NewDownloader(integration.Session)
	params := &s3.GetObjectInput{Bucket: bucketName, Key: &key}

	w := newDLWriter(1024 * 1024 * 20)
	n, err := mgr.Download(w, params)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	if e, a := md5value, fmt.Sprintf("%x", md5.Sum(w.buf[0:n])); e != a {
		t.Errorf("expect %s md5 value, got %s", e, a)
	}
}
