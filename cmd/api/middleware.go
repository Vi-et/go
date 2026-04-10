package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Định nghĩa struct client lưu limiter và thời điểm truy cập cuối.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu sync.Mutex
		// Map giờ trỏ đến con trỏ *client thay vì *rate.Limiter trực tiếp.
		clients = make(map[string]*client)
	)
	// Chạy goroutine nền dọn dẹp các client không còn hoạt động.
	// Goroutine này được khởi tạo 1 lần duy nhất khi middleware được wrap.
	go func() {
		for {
			time.Sleep(time.Minute) // Nghỉ 1 phút rồi mới dọn
			// Lock để ngăn các request kiểm tra limiter trong lúc đang dọn.
			mu.Lock()
			// Xóa mọi client không được thấy trong 3 phút trở lại.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.config.limiter.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		mu.Lock()
		if _, found := clients[ip]; !found {
			// Khởi tạo client mới với rate limiter gắn vào.
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
		}
		// Cập nhật thời điểm truy cập cuối cho IP này.
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
