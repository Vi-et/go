package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	// Thêm điều kiện 'AND version = $6' vào câu truy vấn SQL.
	query := `
        UPDATE movies  
        SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 
        WHERE id = $5 AND version = $6 
        RETURNING version`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // Thêm phiên bản dự kiến của bộ phim.
	}

	// Thực thi câu truy vấn SQL. Nếu không tìm thấy hàng nào khớp, chúng ta biết phiên bản
	// phim đã thay đổi (hoặc bản ghi đã bị xóa) và trả về lỗi ErrEditConflict tùy chỉnh.
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
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

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Sử dụng full-text search cho điều kiện lọc title thay vì so khớp chính xác
	query := fmt.Sprintf(`
    SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version 
    FROM movies 
    WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')  
    AND (genres @> $2 OR $2 = '{}')      
    ORDER BY %s %s, id ASC 
    LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Khởi tạo context với thời lượng giới hạn timeout là 3 giây để tránh treo DB
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Sử dụng QueryContext() để thực thi truy vấn. Kết quả trả về là một sql.Rows
	// Gom tất cả các placeholder vào một slice để dễ quản lý
	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, Metadata{}, err
	}
	// ĐẶC BIỆT QUAN TRỌNG: luôn nhớ dùng defer rows.Close() để giải phóng kết nối
	defer rows.Close()
	// Khởi tạo một mảng lát (slice) trống để chứa trọn bộ dữ liệu phim
	totalRecords := 0
	movies := []*Movie{}

	// Duyệt qua từng dòng dữ liệu bằng rows.Next()
	for rows.Next() {
		// Tạo struct Movie trống để hứng dữ liệu
		var movie Movie
		// Scan ánh xạ từng cột từ DB vào struct.
		// Nhớ sử dụng pq.Array(&movie.Genres) để parser mảng text của PostgreSQL.
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Push struct vừa quét được vào mảng movies
		movies = append(movies, &movie)
	}
	// Khi vòng lặp kết thúc, gọi thêm rows.Err() để đón đầu mọi lỗi phát sinh
	// trong quá trình vòng lặp duyệt dữ liệu.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	// Nếu mọi thứ ổn xuôi, trả về danh sách phim
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return movies, metadata, nil
}
