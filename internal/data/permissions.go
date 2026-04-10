package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

// Định nghĩa Permissions dưới dạng một mảng (slice) các chuỗi, ví dụ:
// ["movies:read", "movies:write"].
type Permissions []string

// Thêm hàm helper Include() để sau này dễ dàng kiểm tra xem mảng Permissions
// có chứa cụ thể một mã quyền nhất định hay không.
func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

// Định nghĩa cấu trúc PermissionModel.
type PermissionModel struct {
	DB *sql.DB
}

// Hàm GetAllForUser() trả về tất cả các mã quyền của một user cụ thể
// dưới dạng mảng Permissions ([]string). Đoạn code này khá quen thuộc
// vì nó dùng pattern trả về nhiều row y hệt như ta đã từng làm.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// Dùng INNER JOIN nối 3 bảng lại với nhau để truy xuất code quyền.
	query := `
        SELECT permissions.code 
        FROM permissions 
        INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id 
        INNER JOIN users ON users_permissions.user_id = users.id 
        WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	// Tác giả dùng cú pháp gộp INSERT và SELECT (kết hợp với 'ANY($2)')
	// Quá trình này sẽ biến mảng truyền vào thành các query song song.
	query := `
        INSERT INTO users_permissions 
        SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Chúng ta dùng hàm pq.Array() của thư viện lib/pq để Go biết cách parse
	// mảng slice thành mảng định dạng text của Postgres
	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
