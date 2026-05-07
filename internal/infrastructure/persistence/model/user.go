// package model

// import (
// 	"time"

// 	"litcart/internal/user/domain"
// )

// type User struct {
// 	ID        int64     `db:"id"`      // 自增主键，仅数据库使用
// 	UserID    int64     `db:"user_id"` // Snowflake ID，对应 Domain.User.ID
// 	Username  string    `db:"username"`
// 	Email     string    `db:"email"`
// 	Password  string    `db:"password"`
// 	Status    int8      `db:"status"`
// 	CreatedAt time.Time `db:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at"`
// }

// // ToDomain 将数据库模型转换为领域实体
// func (m *User) ToDomain() *domain.User {
// 	email, _ := domain.NewEmail(m.Email) // 实际项目中建议更好处理错误

// 	return &domain.User{
// 		ID:        domain.UserID(m.UserID),
// 		Username:  m.Username,
// 		Email:     email,
// 		Status:    domain.UserStatus(m.Status),
// 		CreatedAt: m.CreatedAt,
// 		UpdatedAt: m.UpdatedAt,
// 	}
// }

// // // ToDBModel 将领域实体转换为数据库模型
// //
// //	func (u *domain.User) ToDBModel(passwordHash string) *User {
// //		return &User{
// //			UserID:    u.ID.Int64(),
// //			Username:  u.Username,
// //			Email:     u.Email.String(),
// //			Password:  passwordHash,
// //			Status:    int8(u.Status),
// //			CreatedAt: u.CreatedAt,
// //			UpdatedAt: u.UpdatedAt,
// //		}
// //	}
// //
// // ToDBModel : Domain Entity → DB Model （推荐写法：普通函数）
//
//	func ToDBModel(u *domain.User, passwordHash string) *User {
//		return &User{
//			UserID:    u.ID.Int64(),
//			Username:  u.Username,
//			Email:     u.Email.String(),
//			Password:  passwordHash,
//			Status:    int8(u.Status),
//			CreatedAt: u.CreatedAt,
//			UpdatedAt: u.UpdatedAt,
//		}
//	}
package model

import (
	"time"

	"litcart/internal/user/domain"
)

// User 是数据库行模型,与 domain.User 一一对应但有自己的字段(如自增主键 ID)。
//
// 双 ID 设计:
//   - ID:     InnoDB 自增主键,保证 B+ 树叶子节点顺序写入,提升插入性能
//   - UserID: 业务层使用的 Snowflake ID,对应 domain.User.ID
type User struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Status    int8      `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain DB Model → Domain Entity。
// 这里用 MustNewEmail 风格:DB 中的 email 必然合法(写入时校验过),
// 不合法说明数据损坏,直接 panic 暴露问题。
func (m *User) ToDomain() *domain.User {
	email, err := domain.NewEmail(m.Email)
	if err != nil {
		panic("model: corrupted email in db: " + m.Email)
	}
	return &domain.User{
		ID:           domain.UserID(m.UserID),
		Username:     m.Username,
		Email:        email,
		PasswordHash: m.Password,
		Status:       domain.UserStatus(m.Status),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// FromDomain Domain Entity → DB Model。
// 自增主键 ID 由数据库填,这里不设置。
func FromDomain(u *domain.User) *User {
	return &User{
		UserID:    u.ID.Int64(),
		Username:  u.Username,
		Email:     u.Email.String(),
		Password:  u.PasswordHash,
		Status:    int8(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
