package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/validator"
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

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := parts[1]
		v := validator.New()
		// Validate độ dài chuẩn của token
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Truy xuất User từ Database bằng token với Scope Authentication
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// Nếu hợp lệ, gán cục dữ liệu User thật vào Context
		r = app.contextSetUser(r, user)
		// Trao quyền đi tiếp trong chuỗi handler
		next.ServeHTTP(w, r)

	})

}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lấy thông tin user hiện tại từ context của request
		user := app.contextGetUser(r)

		// Nếu đây là người dùng ẩn danh (nghĩa là không truyền token hợp lệ)
		// Gọi helper trả về lỗi 401 và dùng return để chặn request đi tiếp
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		// Nếu hợp lệ, gọi handler tiếp theo trong chuỗi
		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser kiểm tra xem người dùng đã được đăng nhập và TÀI KHOẢN ĐÃ KÍCH HOẠT chưa.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Thay vì trả về http.HandlerFunc trực tiếp, chúng ta gán quy trình này cho biến `fn`.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// Kiểm tra cờ Activated của user
		// Nếu user chưa kích hoạt, trả về lỗi 403
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	// Bao bọc `fn` bằng middleware `requireAuthenticatedUser`.
	// Điều này đồng nghĩa với việc request đi qua `requireAuthenticatedUser` TRƯỚC
	// sau đó mới đến bước kiểm tra kích hoạt.
	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// 1. Lấy thông tin user hiện tại từ context
		user := app.contextGetUser(r)

		// 2. Gọi CSDL để đem về mảng các quyền của User
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// 3. Hàm Include() mượn từ chương trước. NẾU MẢNG KHÔNG CÓ MÃ NÀY -> Quăng lỗi 403
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Nếu có thì cho đi tiếp
		next.ServeHTTP(w, r)
	}

	// Và điều tuyệt vời nhất: ta lại bọc "kiểm tra quyền" bằng "kiểm tra kích hoạt"
	// -> Luồng bảo vệ 3 tầng hoàn chỉnh được ra đời tự động!
	return app.requireActivatedUser(fn)
}

// Thêm hàm rẽ nhánh này vào dưới cùng hoặc đâu đó trong file
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		// Thêm "Vary: Access-Control-Request-Method" để bọn hệ thống Cache nhận biết
		// Sự phản hồi từ Server sẽ khác nhau phụ thuộc vào lúc khách hỏi Method nào.
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// BẪY LÕNG PREFLIGHT
					// Kiểm tra xem đây có phải là cái bóng "Dò đường" của trình duyệt không
					// Dấu hiệu nhận biết: Dùng method OPTIONS, và Xin phép gửi Method kia.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

						// Xác nhận bật đèn xanh! Server đồng ý thông qua các phương thức nguy hiểm:
						// Lưu ý: Các Method cơ bản như GET, POST, HEAD thì mặc định là pass không cần liệt kê.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						// Mở cổng cho đính kèm thông tin chứng thực và gói khối JSON:
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// "Dò đường" thành công! Chỉ vớt gói tin 200 OK trả luôn tại đây
						// KHÔNG CẦN CHỚM TỚI LOGIC CỦA ROUTER BÊN DƯỚI (next.ServeHTTP).
						// (Trả 200 OK thay vì 204 No Content là để tương thích ngược một số bản trình duyệt cổ)
						w.WriteHeader(http.StatusOK)
						return
					}

					break // Thoát sớm vòng lặp Origin, thả cho Next đi.
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
