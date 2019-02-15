package auth_test

import (
	"github.com/aimeelaplant/comiccruncher/auth"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/auth"
	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPGTokenRepository(t *testing.T) {
	assert.NotNil(t, auth.NewRedisTokenRepository(&redis.Client{}))
}

func TestNewToken(t *testing.T) {
	tk := auth.NewToken("dfsd", "dsfsdf-45")
	assert.NotNil(t, tk)
	assert.Equal(t, "dfsd", tk.Payload)
	assert.Equal(t, "dsfsdf-45", tk.UUID)
}

func TestPGTokenRepositoryCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	o := mock_auth.NewMockRedis(ctrl)
	o.EXPECT().HMSet(gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd())
	tr := auth.NewRedisTokenRepository(o)
	err := tr.Create(&auth.Token{})
	assert.Nil(t, err)
}
