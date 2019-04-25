package imaging

import (
	"bytes"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/disintegration/imaging"
	"go.uber.org/zap"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"time"
)

// Width defines the width of the thumbnail.
type Width int

// Height defines the height of the thumbnail.
type Height int

// Thumbnailer is the interface for generating thumbnails.
type Thumbnailer interface {
	Resize(body io.Reader, width, height int) (*bytes.Buffer, error)
}

// ThumbnailUploader is for uploading generating and uploading thumbnails.
type ThumbnailUploader interface {
	Generate(key string, opts *ThumbnailOptions) ([]*ThumbnailResult, error)
}

// ThumbnailResult is the resulting thumbnail generated.
type ThumbnailResult struct {
	Pathname   string
	Dimensions *ThumbnailSize
}

// ThumbnailSize defines the width and height of the thumbnails.
type ThumbnailSize struct {
	Width  Width
	Height Height
}

// ThumbnailOptions defines options for generating thumbnails.
type ThumbnailOptions struct {
	Sizes          []*ThumbnailSize         // The widths and heights for the thumbnails.
	NamingStrategy storage.FileNameStrategy // The strategy for generating thumbnail filenames.
	RemoteDir      string                   // The remote directory to upload the thumbnails.
}

// InMemoryThumbnailer is for resizing images to thumbnails in memory (doesn't save to  temp file, etc).
type InMemoryThumbnailer struct{}

// Resize resizes the image to the given width and height.
func (t *InMemoryThumbnailer) Resize(body io.Reader, width, height int) (*bytes.Buffer, error) {
	src, _, err := image.Decode(body)
	if err != nil {
		return nil, err
	}
	img := imaging.Resize(src, width, height, imaging.Box)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	return buf, err
}

// S3ThumbnailUploader is for generating and uploading thumbnails to the remote storage.
type S3ThumbnailUploader struct {
	s           storage.Storage
	thumbnailer Thumbnailer
}

// Generate generates thumbnails for the given key stored in S3.
func (u *S3ThumbnailUploader) Generate(key string, opts *ThumbnailOptions) ([]*ThumbnailResult, error) {
	rdr, err := u.s.Download(key)
	defer closeReader(rdr)
	if err != nil {
		return nil, err
	}
	sizes := opts.Sizes
	remoteDir := opts.RemoteDir
	strat := opts.NamingStrategy
	results := make([]*ThumbnailResult, len(sizes))
	for i, size := range sizes {
		// reset seeker so it gets read again
		if i != 0 {
			_, err = rdr.Seek(0, 0)
			if err != nil {
				return nil, err
			}
		}
		buf, err := u.thumbnailer.Resize(rdr, int(size.Width), int(size.Height))
		if err != nil {
			closeReader(buf)
			return nil, err
		}
		filename := remoteDir + strat(string(time.Now().Nanosecond())+".jpg")
		err = u.s.UploadBytes(buf, filename)
		if err != nil {
			closeReader(buf)
			return nil, err
		}
		closeReader(buf)
		results[i] = &ThumbnailResult{Dimensions: size, Pathname: filename}
		log.IMAGING().Info("uploaded thumbnail", zap.String("pathname", filename))
	}
	return results, nil
}

func closeReader(r io.Reader) {
	if r != nil {
		ioutil.NopCloser(r).Close()
	}
}

// NewInMemoryThumbnailer returns a new image thumbnailer.
func NewInMemoryThumbnailer() *InMemoryThumbnailer {
	return &InMemoryThumbnailer{}
}

// NewS3ThumbnailUploader creates a new s3 thumbnail uploader.
func NewS3ThumbnailUploader(s storage.Storage, thumbnailer Thumbnailer) *S3ThumbnailUploader {
	return &S3ThumbnailUploader{
		s:           s,
		thumbnailer: thumbnailer,
	}
}

// NewDefaultThumbnailOptions conveniently creates thumbnail options from the parameters
// with the crc32 time naming strategy as the default filename strat.
func NewDefaultThumbnailOptions(remoteDir string, sizes ...*ThumbnailSize) *ThumbnailOptions {
	szs := make([]*ThumbnailSize, len(sizes))
	for i, size := range sizes {
		szs[i] = size
	}
	return &ThumbnailOptions{
		Sizes:          szs,
		NamingStrategy: storage.Crc32TimeNamingStrategy(),
		RemoteDir:      remoteDir,
	}
}

// NewThumbnailSize creates a new thumbnail size for the given width and height.
func NewThumbnailSize(width, height int) *ThumbnailSize {
	return &ThumbnailSize{Width: Width(width), Height: Height(height)}
}
