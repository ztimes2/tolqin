package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/psqlutil"

	sq "github.com/Masterminds/squirrel"
)

type UserStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{
		db:      db,
		builder: psqlutil.NewQueryBuilder(),
	}
}

func (us *UserStore) UserByEmail(email string) (auth.User, error) {
	query, args, err := us.builder.
		Select("id", "email", "role", "password_hash", "password_salt", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()
	if err != nil {
		return auth.User{}, fmt.Errorf("failed to build query: %w", err)
	}

	var u user
	if err := us.db.QueryRowx(query, args...).StructScan(&u); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, fmt.Errorf("faile to execute query: %w", err)
	}

	return auth.User{
		ID:           u.ID,
		Email:        u.Email,
		Role:         auth.NewRole(u.Role),
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (us *UserStore) CreateUser(e auth.UserCreationEntry) (auth.User, error) {
	query, args, err := us.builder.
		Insert("users").
		Columns("email", "role", "password_hash", "password_salt").
		Values(
			e.Email,
			e.Role.String(),
			e.PasswordHash,
			e.PasswordSalt,
		).
		Suffix("RETURNING id, email, role, password_hash, password_salt, created_at").
		ToSql()
	if err != nil {
		return auth.User{}, fmt.Errorf("failed to build query: %w", err)
	}

	var u user
	if err := us.db.QueryRowx(query, args...).StructScan(&u); err != nil {
		return auth.User{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return auth.User{
		ID:           u.ID,
		Email:        u.Email,
		Role:         auth.NewRole(u.Role),
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		CreatedAt:    u.CreatedAt,
	}, nil
}

type user struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	Role         string    `db:"role"`
	PasswordHash string    `db:"password_hash"`
	PasswordSalt string    `db:"password_salt"`
	CreatedAt    time.Time `db:"created_at"`
}
