package storage_test

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/storage"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestNewS3Storage(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	h := mock_storage.NewMockHttpClient(c)
	h.EXPECT().Get(gomock.Any()).Times(0)
	s3 := mock_storage.NewMockS3Client(c)
	s3.EXPECT().PutObject(gomock.Any()).Times(0)
	s := storage.NewS3Storage(h, s3, nil, "myBucket")
	assert.NotNil(t, s)
}

func TestS3StorageUploadFromRemote(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	file, err := os.Open("testdata/test.png")
	if err != nil {
		panic(err)
	}
	h := mock_storage.NewMockHttpClient(c)
	h.EXPECT().Get(gomock.Any()).Times(1).Return(&http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       file,
	}, nil)
	s3 := mock_storage.NewMockS3Client(c)
	s3.EXPECT().PutObject(gomock.Any()).Times(1).Return(nil, nil)
	s := storage.NewS3Storage(h, s3, nil, "myBucket")
	ui, err := s.UploadFromRemote("test", "/characters/images")
	assert.Nil(t, err)
	assert.NotEmpty(t, ui.MD5Hash)
	assert.NotEmpty(t, ui.Pathname)
}

func TestS3StorageUploadFromRemoteFailsRemoteCall(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	file, err := os.Open("testdata/test.png")
	if err != nil {
		panic(err)
	}
	h := mock_storage.NewMockHttpClient(c)
	h.EXPECT().Get(gomock.Any()).Times(1).Return(&http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       file,
	}, nil)
	s3 := mock_storage.NewMockS3Client(c)
	s3.EXPECT().PutObject(gomock.Any()).Times(1).Return(nil, errors.New("s3 error"))
	s := storage.NewS3Storage(h, s3, nil, "myBucket")
	_, err = s.UploadFromRemote("test", "/characters/images")
	assert.Error(t, err)
}

func TestS3StorageUploadFromRemoteFailsS3Call(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	h := mock_storage.NewMockHttpClient(c)
	h.EXPECT().Get(gomock.Any()).Times(1).Return(&http.Response{
		StatusCode: http.StatusBadGateway,
	}, nil)
	s3 := mock_storage.NewMockS3Client(c)
	s3.EXPECT().PutObject(gomock.Any()).Times(0)
	s := storage.NewS3Storage(h, s3, nil, "myBucket")
	_, err := s.UploadFromRemote("test", "/characters/images")
	assert.Error(t, err)
}

func TestS3StorageDownload(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	d := mock_storage.NewMockS3Downloader(c)
	buf := aws.NewWriteAtBuffer(nil)
	d.EXPECT().Download(buf, gomock.Any()).Return(int64(100), nil)
	st := storage.NewS3Storage(nil, nil, d, "testBucket")
	result, err := st.Download("mykey.jpg")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestS3StorageDownloadError(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	d := mock_storage.NewMockS3Downloader(c)
	buf := aws.NewWriteAtBuffer(nil)
	d.EXPECT().Download(buf, gomock.Any()).Return(int64(0), errors.New("error"))
	st := storage.NewS3Storage(nil, nil, d, "testBucket")
	result, err := st.Download("mykey.jpg")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCrc32TimeNamingStrategy(t *testing.T) {
	assert.True(t, strings.HasSuffix(storage.Crc32TimeNamingStrategy()("test.txt"), ".txt"))
}
