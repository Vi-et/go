package main

import (
	"fmt"
	"net/http"

	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Định nghĩa một struct để hứng dữ liệu từ người dùng gửi lên
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// 2. Sử dụng json.NewDecoder để giải mã Body của Request vào struct 'input'
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Copy các giá trị từ input sang struct Movie
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Khởi tạo một Validator mới
	v := validator.New()

	// Gọi ValidateMovie, nếu không hợp lệ thì ném lỗi 422
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 3. In ra để kiểm tra xem đã nhận đúng chưa
	fmt.Fprintf(w, "%+v\n", input)
}
