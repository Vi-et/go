package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer luôn chạy khi goroutine kết thúc — kể cả khi có panic.
		defer func() {
			// recover() kiểm tra xem có panic đang xảy ra không.
			if err := recover(); err != nil {
				// Set header "Connection: close" — báo cho Go's HTTP server
				// tự động đóng kết nối sau khi response được gửi.
				w.Header().Set("Connection", "close")

				// recover() trả về interface{}, dùng fmt.Errorf để
				// chuyển thành error rồi gọi serverErrorResponse như bình thường.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
