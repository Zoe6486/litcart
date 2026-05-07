// package domain

// import "fmt"

// type UserID int64

// func NewUserID() UserID {
// 	// TODO: 替换为真正的 Snowflake 生成器（推荐 bwmarrin/snowflake）
// 	return UserID(1000000000000000000) // 临时
// }

// func ParseUserID(s string) (UserID, error) {
// 	var id UserID
// 	_, err := fmt.Sscanf(s, "%d", &id)
// 	if err != nil || id <= 0 {
// 		return 0, ErrInvalidID
// 	}
// 	return id, nil
// }

// func (id UserID) String() string { return fmt.Sprintf("%d", id) }
// func (id UserID) Int64() int64   { return int64(id) }
package domain

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
)

type UserID int64

// node 是 package 级 snowflake 节点,通过 InitIDGenerator 在程序启动时注入。
// 不允许 nil node 时调用 NewUserID,会 panic——这是程序员错误,应当尽早暴露。
var node *snowflake.Node

// InitIDGenerator 在 main 启动时调用一次。nodeID 取值范围 [0, 1023]。
func InitIDGenerator(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return err
	}
	node = n
	return nil
}

// NewUserID 生成新的 Snowflake UserID。必须先调用 InitIDGenerator。
func NewUserID() UserID {
	if node == nil {
		panic("domain: id generator not initialized, call InitIDGenerator first")
	}
	return UserID(node.Generate().Int64())
}

// ParseUserID 从字符串解析 UserID。
func ParseUserID(s string) (UserID, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, ErrInvalidID
	}
	return UserID(n), nil
}

func (id UserID) String() string { return strconv.FormatInt(int64(id), 10) }
func (id UserID) Int64() int64   { return int64(id) }
