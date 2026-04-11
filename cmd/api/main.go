package main

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"greenlight.example.com/internal/config"
	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/jsonlog"
	"greenlight.example.com/internal/mailer"
)

// Khai báo hằng số version của ứng dụng
var (
	version   string
	buildTime string
)

// Struct application sẽ chứa các phần phụ thuộc mà các handler cần dùng.
type application struct {
	config config.Config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func openDB(cfg config.Config) (*sql.DB, error) {
	// Gọi sql.Open() - Hàm này TẠO Pool chứ chưa thực sự kết nối DB.
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// Thiết lập đếm ngược trong vòng 5 giây.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Dùng PingContext để test kết nối THỰC TẾ với CSDL. Nếu 5 giây mà văng -> báo lỗi.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Trả về luồng pool database
	return db, nil
}

func main() {

	// Lấy các biến truyền vào config
	// port
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Khởi tạo logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Kết nối database
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Đảm bảo rằng kết nối sẽ được đóng khi hàm main đóng.
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// Publish phiên bản hiện tại ("version") tĩnh dưới dạng chuỗi chữ
	expvar.NewString("version").Set(version)
	// Publish số lượng goroutines đang chạy rầm rộ lúc này
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	// Publish số liệu thống kê realtime về tình trạng connection pool của database
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	// Publish mã Unix timestamp để dễ check thời gian phản hồi server
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Khởi tạo application
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}
