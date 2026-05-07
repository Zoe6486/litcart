package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"litcart/internal/infrastructure/persistence/model"
	"litcart/internal/user/domain"
)

const mysqlErrDupEntry uint16 = 1062

type userRepository struct {
	db *sqlx.DB
}

var _ domain.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create 插入新用户。
// 不传 created_at / updated_at,让 MySQL DEFAULT CURRENT_TIMESTAMP 生效;
// 然后回查这两列填回 domain 实体,保证调用方拿到的实体是完整的。
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	m := model.FromDomain(user)

	const insert = `
		INSERT INTO user (user_id, username, email, password, status, email_verified)
		VALUES (:user_id, :username, :email, :password, :status, :email_verified)`

	if _, err := r.db.NamedExecContext(ctx, insert, m); err != nil {
		return mapInsertError(err)
	}

	// 回查 DB 填充的 created_at / updated_at
	var ts struct {
		CreatedAt sql.NullTime `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}
	const selectTS = `SELECT created_at, updated_at FROM user WHERE user_id = ?`
	if err := r.db.GetContext(ctx, &ts, selectTS, user.ID.Int64()); err == nil {
		if ts.CreatedAt.Valid {
			user.CreatedAt = ts.CreatedAt.Time
		}
		if ts.UpdatedAt.Valid {
			user.UpdatedAt = ts.UpdatedAt.Time
		}
	}
	return nil
}

func (r *userRepository) GetByUserID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, email_verified, created_at, updated_at
		 FROM user WHERE user_id = ? AND status != ?`,
		id.Int64(), domain.StatusDeleted)
}

func (r *userRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, email_verified, created_at, updated_at
		 FROM user WHERE email = ? AND status != ?`,
		email.String(), domain.StatusDeleted)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, email_verified, created_at, updated_at
		 FROM user WHERE username = ? AND status != ?`,
		username, domain.StatusDeleted)
}

func (r *userRepository) UpdatePassword(ctx context.Context, id domain.UserID, passwordHash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE user SET password = ? WHERE user_id = ?`,
		passwordHash, id.Int64())
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return checkAffected(res)
}

func (r *userRepository) UpdateEmailVerified(ctx context.Context, id domain.UserID, verified bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE user SET email_verified = ? WHERE user_id = ?`,
		verified, id.Int64())
	if err != nil {
		return fmt.Errorf("update email verified: %w", err)
	}
	return checkAffected(res)
}

func (r *userRepository) getOne(ctx context.Context, query string, args ...any) (*domain.User, error) {
	var m model.User
	err := r.db.GetContext(ctx, &m, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return m.ToDomain(), nil
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func mapInsertError(err error) error {
	var mErr *mysql.MySQLError
	if !errors.As(err, &mErr) || mErr.Number != mysqlErrDupEntry {
		return fmt.Errorf("insert user: %w", err)
	}
	switch {
	case strings.Contains(mErr.Message, "uidx_username"):
		return domain.ErrUsernameExists
	case strings.Contains(mErr.Message, "uidx_email"):
		return domain.ErrEmailExists
	default:
		return domain.ErrDuplicateEntry
	}
}
