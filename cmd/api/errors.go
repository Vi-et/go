package main

import (
	"fmt"
	"net/http"
)

// logError() là helper dùng để ghi log lỗi. Hiện tại nó chỉ in ra logger,
// nhưng sau này sẽ được nâng cấp lên structured logging.
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// errorResponse() là helper chung để gửi thông báo lỗi dạng JSON.
// Chúng ta sử dụng interface{} cho tham số message để có thể linh hoạt
// truyền vào chuỗi hoặc các cấu trúc phức tạp hơn.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse() dùng khi ứng dụng gặp vấn đề không mong muốn
// ở thời gian chạy (lỗi 500). Nó sẽ ghi log lỗi chi tiết.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// notFoundResponse() dùng để gửi lỗi 404 (Không tìm thấy tài nguyên).
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse() dùng để gửi lỗi 405 (Phương thức không được hỗ trợ).
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// Tác giả tận dụng luôn map[string]string từ package Validator để ném vào báo cáo lỗi JSON
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	// Trả về thông báo lỗi bằng chuỗi dễ hiểu bằng tiếng Anh như định dạng của app
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}
