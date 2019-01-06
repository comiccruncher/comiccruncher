package auth

import (
	"github.com/go-pg/pg/orm"
	"time"
)

// ORM is the interface for interacting with the ORM.
type ORM interface {
	Model(model ...interface{}) *orm.Query
}

// Token is the struct for details about an issued JWT token.
type Token struct  {
	tableName	struct{}   `pg:",discard_unknown_columns"`
	ID 			uint
	Payload 	string
	UUID 		string
	CreatedAt   time.Time  `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt   time.Time  `sql:",notnull,default:NOW()" json:"-"`
}

// PGTokenRepository is the token repository.
type PGTokenRepository struct {
	db ORM
}

// Create persists a new token to the database.
func (r *PGTokenRepository) Create(t *Token) error {
	_, err := r.db.Model(t).Insert()
	return err
}

// NewToken creates a new token struct
func NewToken(payload string, UUID string) *Token {
	return &Token{Payload: payload, UUID: UUID}
}

// NewPGTokenRepository creates a new repository.
func NewPGTokenRepository(orm ORM) *PGTokenRepository {
	return &PGTokenRepository{db: orm}
}
