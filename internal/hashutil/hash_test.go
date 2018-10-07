package hashutil_test

import (
	"github.com/aimeelaplant/comiccruncher/internal/hashutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMD5Hash(t *testing.T) {
	file, err := os.Open("./testdata/test.png")
	defer file.Close()
	md5, err := hashutil.MD5Hash(file)
	assert.Nil(t, err)
	assert.Equal(t, "b9cc76915e5c8a1b007393dae219bd76", md5)
}
