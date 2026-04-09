package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"greenlight.example.com/internal/validator"
)

type MovieModel struct {
	DB *sql.DB
}

// Hàm dùng để thêm mới (Create)
func (m MovieModel) Insert(movie *Movie) error {
	// Thêm Context timeout 3 giây bảo vệ server giống Cách 2
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
        INSERT INTO movies (title, year, runtime, genres) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id, created_at, version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Dùng QueryRowContext thay vì QueryRow giống Cách 1
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Bạn nhớ thêm "errors" vào khối import ở đầu file nhé!

// Hàm dùng để lấy dữ liệu (Read/Get)
func (m MovieModel) Get(id int64) (*Movie, error) {
	// ID của PostgreSQL kiểu bigserial bắt đầu từ 1.
	// Nếu user truyền vào id < 1 thì ta có thể trả về lỗi Not Found luôn đỡ tốn công gọi DB.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// 1. Viết câu SQL SELECT
	query := `
		SELECT id, created_at, title, year, runtime, genres, version 
		FROM movies 
		WHERE id = $1`

	var movie Movie

	// Mách nhỏ: Giống như lúc Insert bạn tự thêm Context timeout cho xịn,
	// Ở đây mình cũng nên làm vậy thay vì gọi m.DB.QueryRow() chay như sách!
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 2. Chạy Query lấy 1 dòng và Scan vào struct
	// LƯU Ý: Vẫn phải dùng pq.Array(&movie.Genres) để dịch cấu trúc Array sang Go slice
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// 3. Xử lý lỗi nếu không tìm thấy phim
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// Bắt lỗi không tìm thấy dòng nào của package sql và chuyển thành lỗi riêng của app
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Hàm dùng để cập nhật (Update)
func (m MovieModel) Update(movie *Movie) error {
	// Lệnh UPDATE SQL: Nhớ cộng version lên 1 mỗi lần thao tác
	// RETURNING version để lấy ra version number mới nhất sau khi sửa
	query := `
		UPDATE movies  
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 
		WHERE id = $5 
		RETURNING version`

	// Chú ý biến $5 map với id của bộ phim
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Thực thi và Scan ghi ngược giá trị version mới vào struct
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
}

// Hàm dùng để xóa (Delete)
func (m MovieModel) Delete(id int64) error {
	// Trả về lỗi luôn nếu ID nhỏ hơn 1 (không hợp lệ)
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies 
		WHERE id = $1`

	// Context Timeout như cũ để phòng hờ kẹt DB
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Thực thi câu query DELETE thông qua ExecContext
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Lấy số lượng bản ghi đã bị xóa
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Nếu bằng 0 thì tức là phim đó không tồn tại
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
