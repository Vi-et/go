package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"greenlight.example.com/internal/data"
)

// Khai báo hằng số version của ứng dụng
const version = "1.0.0"

// Struct config chứa các thiết lập cấu hình. Hiện tại mới chỉ có port và env.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Struct application sẽ chứa các phần phụ thuộc mà các handler cần dùng.
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func openDB(cfg config) (*sql.DB, error) {
	// Gọi sql.Open() - Hàm này TẠO Pool chứ chưa thực sự kết nối DB.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
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
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	// Khởi tạo logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Đảm bảo rằng kết nối sẽ được đóng khi hàm main đóng.
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Thay đổi dòng này:
	logger.Printf("starting %s server on %s", app.config.env, srv.Addr)

	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}

}
