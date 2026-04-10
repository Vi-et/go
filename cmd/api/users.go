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

	// Sử dụng helper background để bao bọc anonymous function thực thi nghiệp vụ gửi mail.
	app.background(func() { 
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user) 
		if err != nil { 
			// Lưu ý KHÔNG DÙNG app.serverErrorResponse ở đây nữa! Thay vào đó chỉ ghi log lại:
			app.logger.PrintError(err, nil) 
		} 
	}) 

	// Nếu vạn sự trót lọt, API xuất format JSON mới ra để báo mã 202 Accepted
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
