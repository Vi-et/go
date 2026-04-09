package main

import (
	"fmt"
	"net/http"

	"greenlight.example.com/internal/data"
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
		app.logger.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 3. In ra để kiểm tra xem đã nhận đúng chưa
	fmt.Fprintf(w, "%+v\n", input)
}
