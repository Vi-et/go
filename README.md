# Greenlight API

**(Go · HTTP Router · PostgreSQL · Production-Ready REST API)**

Lấy cảm hứng và áp dụng các tiêu chuẩn kiến trúc từ Let's Go Further, đây là một hệ thống backend API chuẩn mực, hướng đến tính **dễ bảo trì (maintainable)**, **dễ quan sát (observable)** và **hoạt động ổn định (reliable)**. Hệ thống này tuân thủ các nguyên tắc phát triển backend nghiêm ngặt, chú trọng vào cấu trúc phân tầng, quản lý lỗi tập trung, cấu hình nhất quán và độ an toàn dữ liệu cao.

---

## 1. Core Architecture Doctrine

Hệ thống được thiết kế theo kiến trúc phân tầng rõ ràng (Layered Architecture), tuân thủ chặt chẽ việc tách biệt phần xử lý HTTP và logic nghiệp vụ.

```
Routes (httprouter) → Handlers (Controller) → Models (Repository/Service) → PostgreSQL Database
```

* **Không nhảy vọt lớp (No layer skipping)**: Handlers không gọi trực tiếp SQL truy vấn, mà phải thông qua lớp Models. 
* **Không nhúng Logic (No business logic in routes)**: Config Routes chỉ chịu trách nhiệm định tuyến, bàn giao dữ liệu xử lý cho Handlers chuyên trách.
* **Controllers điều hướng, Models xử lý**: Handlers chịu trách nhiệm phân giải request, gọi model và xử lý format response trả về. Models encapsulate logic database (PostgreSQL connection pool/transactions).

---

## 2. Các nguyên tắc Backend nổi bật (Backend Guidelines Enforced)

### 2.1. Cấu hình tập trung (Centralized Configuration)
Cấu hình không bao giờ bị sử dụng lẻ tẻ bằng `os.Getenv` ở khắp nơi. Tất cả config (DSN, Port, SMTP, Rate Limiting, CORS) được định nghĩa hội tụ tại `cmd/api/main.go` dưới dạng struct `config{}` và chích (Dependency Injection) duy nhất một lần xuống các handlers thông qua struct `application{}`.

### 2.2. Validation đầu vào nghiêm ngặt
Tất cả Request Body, Query Params được Parse và Validate cực kỳ chặt chẽ ngay khi bước vào lớp handler (tương tự với khái niệm Zod validators). Sử dụng custom validator package `internal/validator` để xử lý kiểm tra (email format, max length, password strength, required fields) với các message rõ ràng nhằm ngăn chặn Bad Request xâm nhập Database.

### 2.3. Lỗi phải bị tóm (Centralized Error Handling / Sentry-like pattern)
* Mọi lỗi xảy ra (Routing errors, Logic errors, DB constraint errors) đều được dẫn về package `cmd/api/errors.go` (giống BaseController Error Handling).
* Không `fmt.Println` hay giấu lỗi (silent failures). Logger được chuẩn hóa với module `internal/jsonlog` sinh ra cấu trúc log JSON có level (INFO, ERROR, FATAL, PANIC) để dễ dàng trace & monitor.
* Trang bị middleware RecoverPanic: Tự động chụp lại các Exception (Panic), capture callstack trace và trả về HTTP 500 đẹp mắt thay vì sập server.

### 2.4. DI (Dependency Injection) Xuyên Suốt
Handlers không import Models một cách trực tiếp từ global variables. Chúng nhận các Depedencies như Logger, Db connection pools (Models), Mailer thông qua pointer của receiver struct `application` method.

### 2.5. Bảo vệ hệ thống (Observability & Operational Safety)
* **Metrics Monitoring:** Sử dụng expvar phơi bày metric về database conns pool, ram memory chạy ngầm để setup Dashboard (BFRI index theo dõi).
* **Graceful Shutdown**: Hệ thống có cơ chế bắt tín hiệu tắt (SIGINT, SIGTERM) và chờ xử lý nốt các in-flight requests hay async queue trước khi ngắt hoàn toàn kết nối xuống Database.
* **Background Processing**: Gửi email xử lý không tuần tự nhằm chống kẹt băng thông chính, thiết lập cùng waitgroup bảo vệ.

---

## 3. Directory Structure Quy Chuẩn

```
let-go-further/
├── cmd/
│   └── api/             # Entrypoint chính (main.go, cấu hình, routes, auth middleware)
│       ├── main.go      # Khởi chạy server, DB Pool, Load Config
│       ├── routes.go    # Express-like router (httprouter)
│       ├── middleware.go# Auth, RateLimiter, CORS, Panic Recover
│       ├── errors.go    # Formatter các loại error trả về (Sentry-like intent)
│       ├── helpers.go   # Read/Write JSON responses, parse URLs
│       ├── *handlers.go # Endpoint Controllers: chia tách theo domain (users, tokens, movies)
├── internal/
│   ├── data/            # (Repository Layer) Models, Database Queries, DB Schemas, Validation 
│   ├── jsonlog/         # Logger format JSON tùy chỉnh
│   ├── mailer/          # Service xử lý gửi SMTP Mail
│   ├── validator/       # Module kiểm tra dữ liệu đầu vào (Zod alternative)
├── migrations/          # Quản lý Database Schemas SQL (UP/DOWN scripts)
├── Makefile             # Quản lý task automation (Build, Run DB, Migrate, Dev)
└── go.mod               # Dependencies
```

---

## 4. Bắt đầu ngay (Getting Started)

Dự án có sẵn cơ chế auto-rebuild với `Make`. Hãy chắc chắn đã cài PostgreSQL và cài đặt các biến môi trường cấu hình bắt buộc.

Khởi động DB và chạy dự án (Server Development):
```bash
make dev
```
*(Server mặc định chạy tại `localhost:4000`)*

## 5. Checklist Trước Khi Thay Đổi (Developer Validation)
Trước khi tạo PR / Commit tính năng:
- [ ] Tuân thủ Layered Architecture? Không nhét Business/DB Logic vào trong `cmd/api/*handlers.go`.
- [ ] Mọi Struct truyền từ client đã được Validate qua `validator.Validator` trong thư mục data chưa?
- [ ] Lỗi được trả qua `app.serverErrorResponse()` hoặc format chuẩn ở `errors.go` thay vì panics?
- [ ] Đã thêm Migration SQL (nếu có cấu trúc DB thay đổi)?
- [ ] Background goroutines (nếu chạy nền) đã gán `app.background()` để bắt wg chưa? 

**Status:** Stable · Enforceable · Production-grade API
