package storage

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/hashutil"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"hash/crc32"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Storage defines the interface for uploading remote files to a remote directory.
type Storage interface {
	Download(key string) (*bytes.Reader, error)
	UploadFromRemote(remoteURL string, remoteDir string) (UploadedImage, error)
	UploadBytes(b *bytes.Buffer, remotePathName string) error
}

// S3Client is the interface for interacting with S3-related stuff.
type S3Client interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

// S3Downloader downloads files from s3.
type S3Downloader interface {
	Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error)
}

// HTTPClient is the interface for interacting with http calls.
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

// S3StorageOption sets optional parameters.
type S3StorageOption func(*S3Storage)

// NamingStrategy sets the naming strategy for the storage.
func NamingStrategy(strategy FileNameStrategy) S3StorageOption {
	return func(s *S3Storage) {
		s.namingStrategy = strategy
	}
}

// S3Storage is the Storage implementation for AWS S3.
type S3Storage struct {
	httpClient     HTTPClient
	s3             S3Client         // The s3 storage.
	s3Downloader   S3Downloader     // The s3 downloader.
	bucket         string           // The name of the S3 bucket.
	namingStrategy FileNameStrategy // The naming strategy for uploading a file to S3.
}

// FileNameStrategy is a callable used for naming a file.
type FileNameStrategy func(basename string) string

// UploadedImage is the uploaded image with its pathname and md5 hash of the image data.
type UploadedImage struct {
	Pathname string
	MD5Hash  string
}

// Download downloads the specified key and returns the bytes.
func (storage *S3Storage) Download(key string) (*bytes.Reader, error) {
	in := &s3.GetObjectInput{
		Bucket: &storage.bucket,
		Key:    &key,
	}
	buf := aws.NewWriteAtBuffer(nil)
	_, err := storage.s3Downloader.Download(buf, in)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// UploadFromRemote uploads a file from a remote url. The remote file gets temporarily read in memory.
func (storage *S3Storage) UploadFromRemote(remoteFile string, remoteDir string) (UploadedImage, error) {
	var uploadImage UploadedImage
	u, err := url.Parse(remoteFile)
	if err != nil {
		return uploadImage, fmt.Errorf("cannot parse url: %s", err)
	}
	res, err := storage.httpClient.Get(remoteFile)
	if res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return uploadImage, fmt.Errorf("error requesting the remote url: %s", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotModified {
		return uploadImage, fmt.Errorf("got bad status code from remote url %s: %d", remoteFile, res.StatusCode)
	}
	// Check if there is not a leading slash in the remoteDir.
	if !strings.HasSuffix(remoteDir, "/") {
		// Add a leading slash. :)
		remoteDir = remoteDir + "/"
	}
	// copy for later.
	b, err := ioutil.ReadAll(res.Body)
	buf := bytes.NewBuffer(b)
	if buf != nil {
		defer ioutil.NopCloser(buf).Close()
	}
	if err != nil {
		return uploadImage, err
	}
	remotePathName := remoteDir + storage.namingStrategy(filepath.Base(u.Path))
	if err := storage.UploadBytes(buf, remotePathName); err != nil {
		return uploadImage, fmt.Errorf("could not upload: %s", err)
	}
	md5Hash, err := hashutil.MD5Hash(buf)
	if err != nil {
		return uploadImage, err
	}
	uploadImage.MD5Hash = md5Hash
	uploadImage.Pathname = remotePathName
	return uploadImage, nil
}

// UploadBytes uploads an item as bytes with the specified name.
func (storage *S3Storage) UploadBytes(b *bytes.Buffer, remotePathName string) error {
	ctx := context.Background()
	timeout := time.Duration(10 * time.Second) // 10 seconds
	_, cancelFn := context.WithTimeout(ctx, timeout)
	defer cancelFn()
	byts := b.Bytes()
	if _, err := storage.s3.PutObject(
		&s3.PutObjectInput{
			Bucket:        aws.String(storage.bucket),
			Body:          bytes.NewReader(byts),
			ContentType:   aws.String(http.DetectContentType(byts)),
			ContentLength: aws.Int64(int64(b.Len())),
			Key:           aws.String(remotePathName),
			CacheControl:  aws.String("max-age=2592000"),
		},
	); err != nil {
		return err
	}
	return nil
}

// NewAwsSessionFromEnv creates a new aws session from environment variables.
func NewAwsSessionFromEnv() (*session.Session, error) {
	creds := credentials.Value{
		AccessKeyID:     os.Getenv("CC_AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("CC_AWS_SECRET_ACCESS_KEY"),
	}
	return session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("CC_AWS_REGION")),
		Credentials: credentials.NewStaticCredentialsFromCreds(creds),
	})
}

// NewS3StorageFromEnv creates the new S3 storage implementation from env vars.
func NewS3StorageFromEnv() (*S3Storage, error) {
	ses, err := NewAwsSessionFromEnv()
	if err != nil {
		return nil, err
	}
	return NewS3Storage(http.DefaultClient, s3.New(ses), s3manager.NewDownloader(ses), os.Getenv("CC_AWS_BUCKET")), nil
}

// Crc32TimeNamingStrategy returns the crc32 encoded string of the unix time in nanoseconds plus the file extension
// of the given basename.
func Crc32TimeNamingStrategy() FileNameStrategy {
	return func(basename string) string {
		// Create a new instance every time to make it concurrent-safe.
		crcHasher := crc32.NewIEEE()
		crcHasher.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
		return hex.EncodeToString(crcHasher.Sum(nil)) + filepath.Ext(basename)
	}
}

// NewS3Storage creates a new S3 storage implementation from params.
func NewS3Storage(httpClient HTTPClient, s3 S3Client, s3Downloader S3Downloader, bucket string, opts ...S3StorageOption) *S3Storage {
	s := &S3Storage{
		httpClient:     httpClient,
		s3:             s3,
		bucket:         bucket,
		s3Downloader:   s3Downloader,
		namingStrategy: Crc32TimeNamingStrategy(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
