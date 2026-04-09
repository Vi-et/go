package data

import (
	"database/sql"
	"errors"
)

// Tạo một lỗi tuỳ chỉnh ErrRecordNotFound
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Struct Models chứa MovieModel (và các model khác sau này)
type Models struct {
	Movies MovieModel
}

// Hàm khởi tạo Models, nhận vào DB pool và truyền vào MovieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
