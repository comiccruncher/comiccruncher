package imaging_test

import (
	"bytes"
	"github.com/comiccruncher/comiccruncher/imaging"
	"github.com/comiccruncher/comiccruncher/internal/mocks/imaging"
	"github.com/comiccruncher/comiccruncher/internal/mocks/storage"
	"github.com/comiccruncher/comiccruncher/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestNewInMemoryThumbnailer(t *testing.T) {
	assert.NotNil(t, imaging.NewInMemoryThumbnailer())
}

func TestNewDefaultThumbnailOptions(t *testing.T) {
	opts := imaging.NewDefaultThumbnailOptions("testy/", imaging.NewThumbnailSize(100, 100), imaging.NewThumbnailSize(200, 200))

	assert.NotNil(t, opts)
	assert.Equal(t, "testy/", opts.RemoteDir)
	assert.Len(t, opts.Sizes, 2)
	assert.IsType(t, storage.Crc32TimeNamingStrategy(), opts.NamingStrategy)
}

func TestNewThumbnailSize(t *testing.T) {
	size := imaging.NewThumbnailSize(100, 100)
	assert.NotNil(t, size)
	assert.Equal(t, 100, int(size.Width))
	assert.Equal(t, 100, int(size.Height))
}

func TestInMemoryThumbnailerResize(t *testing.T) {
	thmb := imaging.NewInMemoryThumbnailer()
	file, err := os.Open("testdata/dummy.jpg")
	must(err)
	defer closeReader(file)
	stat, err := file.Stat()
	must(err)
	originalSize := stat.Size()
	assert.True(t, originalSize > 0)
	buf, err := thmb.Resize(file, 100, 100)
	if buf != nil {
		defer buf.Reset()
	}
	assert.Nil(t, err)
	size := len(buf.Bytes())
	assert.True(t, size > 0)
	assert.True(t, originalSize > int64(size))
}

func TestNewS3ThumbnailUploader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mock_storage.NewMockStorage(ctrl)
	mockThumb := mock_imaging.NewMockThumbnailer(ctrl)
	uploader := imaging.NewS3ThumbnailUploader(mockStorage, mockThumb)
	assert.NotNil(t, uploader)
}

func TestNewS3ThumbnailUploaderGenerate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	file, err := os.Open("testdata/dummy.jpg")
	must(err)
	file2, err := os.Open("testdata/dummy.jpg")
	must(err)
	defer closeReader(file)
	defer closeReader(file2)
	byts, err := ioutil.ReadAll(file)
	must(err)
	byts2, err := ioutil.ReadAll(file)
	must(err)

	rdr := bytes.NewReader(byts)
	buf := bytes.NewBuffer(byts)
	buf2 := bytes.NewBuffer(byts2)

	mockStorage := mock_storage.NewMockStorage(ctrl)
	mockStorage.EXPECT().Download("testy.jpg").Return(rdr, nil)
	mockThumb := mock_imaging.NewMockThumbnailer(ctrl)
	mockThumb.EXPECT().Resize(gomock.Any(), 100, 100).Return(buf, nil)
	mockThumb.EXPECT().Resize(gomock.Any(), 300, 300).Return(buf2, nil)

	mockStorage.EXPECT().UploadBytes(buf, gomock.Any()).Return(nil)
	mockStorage.EXPECT().UploadBytes(buf2, gomock.Any()).Return(nil)

	uploader := imaging.NewS3ThumbnailUploader(mockStorage, mockThumb)

	opts := &imaging.ThumbnailOptions{
		Sizes: []*imaging.ThumbnailSize{
			{Width: 100, Height: 100},
			{Width: 300, Height: 300},
		},
		NamingStrategy: storage.Crc32TimeNamingStrategy(),
		RemoteDir:      "testdir/",
	}

	results, err := uploader.Generate("testy.jpg", opts)
	res1 := results[0]
	res2 := results[1]

	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 100, int(res1.Dimensions.Height))
	assert.Equal(t, 100, int(res1.Dimensions.Width))
	assert.NotEmpty(t, res1.Pathname)
	assert.True(t, strings.HasSuffix(res1.Pathname, ".jpg"))
	assert.Equal(t, 300, int(res2.Dimensions.Height))
	assert.Equal(t, 300, int(res2.Dimensions.Width))
	assert.NotEmpty(t, res2.Pathname)
	assert.True(t, res2.Pathname != res1.Pathname)
	assert.True(t, strings.HasSuffix(res2.Pathname, ".jpg"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func closeReader(r io.ReadCloser) {
	if r != nil {
		r.Close()
	}
}
