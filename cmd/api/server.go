package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Khởi tạo HTTP server với các cài đặt giống như trong main() cũ
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Gọi srv.Shutdown() và truyền context 5 giây vào.
		// Hàm srv.Shutdown sẽ cố tình giữ các kết nối đang dở dang và giải quyết nốt,
		// nó sẽ trả kết quả thành công(nil) nếu dọn xong, hoặc trả lỗi nếu lố qua 5 giây.
		// Gắn lỗi này tống vào channel shutdownError ở phía ngoài.
		shutdownError <- srv.Shutdown(ctx)

	}()

	// Log thông báo "đang khởi động server"
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	// Luồng code sẽ chạy xuống tận đây và bị BLOCK LẠI vì chờ lấy dữ liệu từ shutdownError
	// Đây chính là kỹ thuật giữ chân server không cho nó hẹo, đợi cho goroutine chạy đủ 5 giây
	// Nếu kết quả trả ra từ Shutdown() là lỗi khác nil, ta return.
	err = <-shutdownError
	if err != nil {
		return err
	}
	// Tại tọa độ này, Shutdown() đã kết thúc an toàn thành công nhường lại điều kiện,
	// ta thông báo một câu cuối cùng là "stopped server" rồi nghỉ hưu thật sự.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
