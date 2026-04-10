package main

import (
	"errors"
	"net/http"

	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.SetPassword(input.Password)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateRegister(v, user)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Đẩy bản lưu vào Database bằng cách gọi model
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// Khi lỗi sinh ra là ErrDuplicateEmail, API sẽ tự bồi thêm thông báo validation lỗi giả cảnh báo email đã tồn tại gửi lại cho Client
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Nếu vạn sự trót lọt, API xuất format JSON mới ra để báo mã 201 Created
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
