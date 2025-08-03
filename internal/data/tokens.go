package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"greenlight.samedarslan28.net/internal/validator"
	"time"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(scope string, userID int64, ttl time.Duration) (*Token, error) {
	token := &Token{
		Scope:  scope,
		UserID: userID,
		Expiry: time.Now().Add(ttl),
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (t TokenModel) New(userID int64, ttl time.Duration) (*Token, error) {
	token, err := generateToken(ScopeActivation, userID, ttl)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t TokenModel) Insert(token *Token) error {
	query := `
INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES ($1, $2, $3, $4);
`
	args := []interface{}{token.Plaintext, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (t TokenModel) DeleDeleteAllForUser(scope string, userID int64) error {
	query := `DELETE FROM tokens WHERE user_id = $1 AND scope = $2;`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := t.DB.ExecContext(ctx, query, userID, scope)
	if err != nil {
		return err
	}
	return nil
}
