package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSaltByteSize = 16
)

type PasswordSalter struct {
	byteSize int
	reader   io.Reader
	encodeFn func([]byte) string
}

func NewPasswordSalter() *PasswordSalter {
	return &PasswordSalter{
		reader:   rand.Reader,
		byteSize: defaultSaltByteSize,
		encodeFn: base64.URLEncoding.EncodeToString,
	}
}

func (p *PasswordSalter) GenerateSalt() (string, error) {
	b := make([]byte, p.byteSize)
	if _, err := p.reader.Read(b); err != nil {
		return "", err
	}

	return p.encodeFn(b), nil
}

func (p *PasswordSalter) SaltPassword(password, salt string) string {
	return password + salt
}

var (
	ErrMismatchedPassword = errors.New("mismatched password")
)

type PasswordHasher struct {
	cost              int
	hashPasswordFn    func(password []byte, cost int) ([]byte, error)
	comparePasswordFn func(hash, password []byte) error
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost:              bcrypt.DefaultCost,
		hashPasswordFn:    bcrypt.GenerateFromPassword,
		comparePasswordFn: bcrypt.CompareHashAndPassword,
	}
}

func (p *PasswordHasher) HashPassword(password string) (string, error) {
	hash, err := p.hashPasswordFn([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *PasswordHasher) ComparePassword(hash, password string) error {
	if err := p.comparePasswordFn([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatchedPassword
		}
		return err
	}
	return nil
}

type Tokener struct {
	issuer        string
	signingKey    string
	signingMethod jwt.SigningMethod
	expiry        time.Duration
	timeNowFn     func() time.Time
}

func NewTokener(issuer, signingKey string, expiry time.Duration) *Tokener {
	return &Tokener{
		issuer:        issuer,
		signingKey:    signingKey,
		signingMethod: jwt.SigningMethodHS256,
		expiry:        expiry,
		timeNowFn:     time.Now,
	}
}

func (t *Tokener) Token(u User) (string, error) {
	now := t.timeNowFn()

	claims := TokenClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   u.ID,
			Issuer:    t.issuer,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(t.expiry).Unix(),
		},
		Email: u.Email,
		Role:  u.Role,
	}

	return jwt.NewWithClaims(t.signingMethod, &claims).SignedString([]byte(t.signingKey))
}

func (t *Tokener) ParseTokenClaims(token string) (TokenClaims, error) {
	var claims TokenClaims

	if _, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(t.signingKey), nil
	}); err != nil {
		return TokenClaims{}, err
	}

	return claims, nil
}

type TokenClaims struct {
	jwt.StandardClaims

	Email string `json:"email,omitempty"`
	Role  Role   `json:"role,omitempty"`
}

func (t TokenClaims) Valid() error {
	// TODO validate more claims
	return t.StandardClaims.Valid()
}

type contextKey struct{}

var tokenClaimsKey contextKey = struct{}{}

func ContextWith(ctx context.Context, t TokenClaims) context.Context {
	return context.WithValue(ctx, tokenClaimsKey, t)
}

func FromContext(ctx context.Context) (TokenClaims, bool) {
	t, ok := ctx.Value(tokenClaimsKey).(TokenClaims)
	return t, ok
}

var (
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrNotAuthorized    = errors.New("not authorized")
)

func WithRoleFromContext(ctx context.Context, r Role) (TokenClaims, error) {
	claims, ok := FromContext(ctx)
	if !ok {
		return TokenClaims{}, ErrNotAuthenticated
	}

	if claims.Role != r {
		return TokenClaims{}, ErrNotAuthorized
	}

	return claims, nil
}
