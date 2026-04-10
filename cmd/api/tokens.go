// File: cmd/api/tokens.go
package main

import (
	"errors"
	"net/http"
	"time"

	// Lưu ý: Đổi greenlight.example.com thành module path của bạn (nếu có)
	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePassword(v, input.Password) // Sách dùng hàm ValidatePasswordPlaintext

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 3. Truy xuất user theo email
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// SỬ DỤNG LỖI 401 UNAUTHORIZED THAY VÌ 422
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 4. So sánh khớp mật khẩu
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		// SỬ DỤNG LỖI 401 UNAUTHORIZED
		app.invalidCredentialsResponse(w, r)
		return
	}

	// 5. Cấp Token 24h
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 6. Trả về token với HTTP Status 201 Created và định dạng authentication_token
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
