package hashutil

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

// Returns the MD5 hash of a file. It's the caller's responsibility to close the reader or reset the seek.
func MD5Hash(body io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, body); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}
