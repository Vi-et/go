package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.example.com/internal/validator"
)

var ErrDuplicateEmail = errors.New("duplicate email")

// Định nghĩa struct đại diện cho một thông tin người dùng.
// Quan trọng: Sử dụng tag json:"-" để ngăn trường Password và Version hiển thị mỗi khi xuất ra JSON.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type UserModel struct {
	DB *sql.DB
}

// password struct chứa 2 phiên bản của mật khẩu: chuỗi gốc (dạng con trỏ) và chuỗi mã hoá ([]byte).
// Việc dùng con trỏ cho plaintext giúp phân biệt trường hợp mật khẩu bị null (vô tình thiếu) hay cố tình bỏ trống "".
type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintext
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
	if err != nil {
		return false, err
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 characters long")
	v.Check(len(password) < 72, "password", "must not be longer than 72 characters")

}

func ValidateRegister(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 50, "name", "must not be longer than 50 characters")

	ValidateEmail(v, user.Email)
	ValidatePassword(v, *user.Password.plaintext)
}

// Insert tạo bảng ghi mới và sử dụng RETURNING clause (giống movies) để lấy lại data từ database.
func (m UserModel) Insert(user *User) error {
	query := `
        INSERT INTO users (name, email, password_hash, activated)  
        VALUES ($1, $2, $3, $4) 
        RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Nhờ có cờ Unique, PostgreSQL sẽ trả về thông báo lỗi nếu Email tạo bị trùng
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, activated, version 
        FROM users 
        WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
        UPDATE users  
        SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1 
        WHERE id = $5 AND version = $6 
        RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
