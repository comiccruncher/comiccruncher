package auth_test

import (
	"github.com/aimeelaplant/comiccruncher/auth"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/auth"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPGTokenRepository(t *testing.T) {
	assert.NotNil(t, auth.NewPGTokenRepository(&pg.DB{}))
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

	o := mock_auth.NewMockORM(ctrl)
	o.EXPECT().Model(gomock.Any()).Return(orm.NewQuery(nil, nil))
	tr := auth.NewPGTokenRepository(o)
	err := tr.Create(&auth.Token{})
	assert.NotNil(t, err)
}
