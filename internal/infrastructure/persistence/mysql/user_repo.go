// package mysql

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"strings"

// 	"github.com/go-sql-driver/mysql"
// 	"github.com/jmoiron/sqlx"

// 	"litcart/internal/infrastructure/persistence/model"
// 	"litcart/internal/user/domain"
// )

// type userRepository struct {
// 	db *sqlx.DB
// }

// func NewUserRepository(db *sqlx.DB) domain.UserRepository {
// 	return &userRepository{db: db}
// }

// func (r *userRepository) Create(ctx context.Context, user *domain.User, passwordHash string) error {
// 	// dbModel := user.ToDBModel(passwordHash)
// 	dbModel := model.ToDBModel(user, passwordHash)

// 	query := `INSERT INTO user (user_id, username, email, password, status, created_at, updated_at)
// 			  VALUES (:user_id, :username, :email, :password, :status, :created_at, :updated_at)`

// 	_, err := r.db.NamedExecContext(ctx, query, dbModel)
// 	if err != nil {
// 		// 处理唯一键冲突
// 		var mysqlErr *mysql.MySQLError
// 		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
// 			msg := mysqlErr.Message
// 			switch {
// 			case strings.Contains(msg, "uidx_username"):
// 				return domain.ErrUsernameExists
// 			case strings.Contains(msg, "uidx_email"):
// 				return domain.ErrEmailExists
// 			default:
// 				return err
// 			}
// 		}
// 		return err
// 	}
// 	return nil
// }

// func (r *userRepository) GetByUserID(ctx context.Context, id domain.UserID) (*domain.User, error) {
// 	var m model.User
// 	err := r.db.GetContext(ctx, &m,
// 		`SELECT id, user_id, username, email, password, status, created_at, updated_at
// 		 FROM user WHERE user_id = ? AND status != ?`,
// 		id.Int64(), domain.StatusDeleted)

// 	if err == sql.ErrNoRows {
// 		return nil, domain.ErrUserNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	return m.ToDomain(), nil
// }

// func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
// 	var m model.User
// 	err := r.db.GetContext(ctx, &m,
// 		`SELECT id, user_id, username, email, password, status, created_at, updated_at
// 		 FROM user WHERE email = ? AND status != ?`,
// 		email, domain.StatusDeleted)

// 	if err == sql.ErrNoRows {
// 		return nil, domain.ErrUserNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return m.ToDomain(), nil
// }

// func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
// 	var m model.User
// 	err := r.db.GetContext(ctx, &m,
// 		`SELECT id, user_id, username, email, password, status, created_at, updated_at
// 		 FROM user WHERE username = ? AND status != ?`,
// 		username, domain.StatusDeleted)

// 	if err == sql.ErrNoRows {
// 		return nil, domain.ErrUserNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return m.ToDomain(), nil
// }

// // func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
// // 	dbModel := user.ToDBModel()

// // 	query := `
// //         UPDATE users
// //         SET name = :name, email = :email, status = :status, updated_at = :updated_at
// //         WHERE id = :id`

// // 	_, err := r.db.NamedExecContext(ctx, query, dbModel)
// // 	return err
// // }
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

// 编译期检查:userRepository 必须实现 domain.UserRepository。
// 接口签名变了忘改实现时,这里立刻编译报错。
var _ domain.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	m := model.FromDomain(user)

	const query = `
		INSERT INTO user (user_id, username, email, password, status, created_at, updated_at)
		VALUES (:user_id, :username, :email, :password, :status, :created_at, :updated_at)`

	//如果没接 ctx,数据库查询不受请求超时控制,慢查询会积压。
	if _, err := r.db.NamedExecContext(ctx, query, m); err != nil {
		return mapInsertError(err)
	}
	return nil
}

func (r *userRepository) GetByUserID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, created_at, updated_at
		 FROM user WHERE user_id = ? AND status != ?`,
		id.Int64(), domain.StatusDeleted)
}

func (r *userRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, created_at, updated_at
		 FROM user WHERE email = ? AND status != ?`,
		email.String(), domain.StatusDeleted)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT id, user_id, username, email, password, status, created_at, updated_at
		 FROM user WHERE username = ? AND status != ?`,
		username, domain.StatusDeleted)
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

// mapInsertError 把唯一索引冲突映射为领域错误。
// 索引名约定:uidx_username / uidx_email
// 未识别的唯一索引冲突走 ErrDuplicateEntry 兜底,避免泄露原始 MySQL 错误。
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
