package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Khai báo hằng số version của ứng dụng
const version = "1.0.0"

// Struct config chứa các thiết lập cấu hình. Hiện tại mới chỉ có port và env.
type config struct {
	port int
	env  string
}

// Struct application sẽ chứa các phần phụ thuộc mà các handler cần dùng.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Port to run the server on")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Khởi tạo logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	// 2. Đăng ký endpoint healthcheck
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)
	// Đăng ký endpoint createMovieHandler
	mux.HandleFunc("/v1/movies", app.createMovieHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Thay đổi dòng này:
	logger.Printf("starting %s server on %s", app.config.env, srv.Addr)

	err := srv.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}

}
