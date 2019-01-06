package auth

import (
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPGTokenRepository(t *testing.T) {
	assert.NotNil(t, NewPGTokenRepository(&pg.DB{}))
}

func TestNewToken(t *testing.T) {
	tk := NewToken("dfsd", "dsfsdf-45")
	assert.NotNil(t, tk)
	assert.Equal(t, "dfsd", tk.Payload)
	assert.Equal(t, "dsfsdf-45", tk.UUID)
}
