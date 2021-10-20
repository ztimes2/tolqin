package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
)

const (
	roleNameAdmin = "admin"
)

// RoleName returns the given role's string representation that is used for the
// respective JWT claim. An empty string is returned for unknown roles.
func RoleName(r auth.Role) string {
	switch r {
	case auth.RoleAdmin:
		return roleNameAdmin
	default:
		return ""
	}
}

// Role returns auth.Role for the given role string that is used for the respective
// JWT claim. auth.RoleUndefined is returned for unknown role strings.
func Role(s string) auth.Role {
	switch s {
	case roleNameAdmin:
		return auth.RoleAdmin
	default:
		return auth.RoleUndefined
	}
}

// EncodeDecoder takes care of encoding and decoding the application's JWTs.
type EncodeDecoder struct {
	signingKey    string
	signingMethod jwt.SigningMethod
	expiry        time.Duration
	timeNowFn     func() time.Time
}

// NewEncodeDecoder returns a new *EncodeDecoder using the given singing key and
// the duration until JWTs expire.
func NewEncodeDecoder(signingKey string, expiry time.Duration) *EncodeDecoder {
	return &EncodeDecoder{
		signingKey:    signingKey,
		signingMethod: jwt.SigningMethodHS256,
		expiry:        expiry,
		timeNowFn:     time.Now,
	}
}

// EncodeJWT encodes the given user to JWT.
func (ed *EncodeDecoder) EncodeJWT(u auth.User) (string, error) {
	now := ed.timeNowFn()

	c := Claims{
		StandardClaims: jwt.StandardClaims{
			Subject:   u.ID,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(ed.expiry).Unix(),
		},
		Email: u.Email,
		Role:  RoleName(u.Role),
	}

	return jwt.NewWithClaims(ed.signingMethod, &c).SignedString([]byte(ed.signingKey))
}

// DecodeJWT decodes the given string, validates and returns its JWT claims.
func (ed *EncodeDecoder) DecodeJWT(s string) (Claims, error) {
	var c Claims

	if _, err := jwt.ParseWithClaims(s, &c, func(_ *jwt.Token) (interface{}, error) {
		return []byte(ed.signingKey), nil
	}); err != nil {
		return Claims{}, err
	}

	return c, nil
}

// Claims holds the application's JWT claims.
type Claims struct {
	jwt.StandardClaims

	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}

// Valid validates the given JWT claims.
func (c Claims) Valid() error {
	return c.StandardClaims.Valid()
}

type contextKey struct{}

var claimsKey contextKey = struct{}{}

// ContextWith attaches the given JWT claims to a context.
func ContextWith(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// FromContext retrieves JWT claims from the given context.
func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey).(Claims)
	return c, ok
}

// WithRoleFromContext retrieves JWT claims containing the given role from the
// given context.
//
// ErrClaimsNotFound is returned when the context doesn't contain the expected JWT
// claims. ErrRoleMismatched is returned when the JWT claims doesn't contain the
// expected role.
func WithRoleFromContext(ctx context.Context, r auth.Role) (Claims, error) {
	c, ok := FromContext(ctx)
	if !ok {
		return Claims{}, ErrClaimsNotFound
	}

	if Role(c.Role) != r {
		return Claims{}, ErrMismatchedRole
	}

	return c, nil
}

var (
	// ErrClaimsNotFound is used when JWT claims could not be found.
	ErrClaimsNotFound = errors.New("jwt claims not found")

	// ErrMismatchedRole is used when JWT role claim is mismatched.
	ErrMismatchedRole = errors.New("mismatched jwt role claim")
)
