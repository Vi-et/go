include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: in ra thông báo trợ giúp này
.PHONY: help
help:
	@echo 'Cách dùng:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | sed -e 's/^/ /'

# ==================================================================================== #
# DATABASE MIGRATIONS
# ==================================================================================== #

## db/migrations/new name=$1: tạo một file migration mới
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Đang tạo các file migration cho ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: chạy tất cả các migration "up"
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Đang chạy các migration up...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

## db/migrations/down: chạy tất cả các migration "down"
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Đang chạy các migration down...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} down

## dev: chạy ứng dụng bằng air và truyền tham số từ .env
.PHONY: dev
dev:
	@go tool air -- -db-dsn=${GREENLIGHT_DB_DSN} \
		-limiter-rps=${LIMITER_RPS} \
		-limiter-burst=${LIMITER_BURST} \
		-limiter-enabled=${LIMITER_ENABLED}

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...


.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# Lấy thời gian hiện tại ở định dạng ISO-8601
current_time = $(shell date --iso-8601=seconds)
# Lấy mô tả trạng thái repo từ Git (bao gồm tag, số commit, hash)
# --always: luôn trả về giá trị dù chưa có tag
# --dirty: thêm hậu tố "-dirty" nếu có file chưa commit
# --tags: ưu tiên dùng tag nếu có
# --long: luôn hiển thị đầy đủ {tag}-{n_commits}-g{hash}
git_description = $(shell git describe --always --dirty --tags --long)
# Gom linker flags vào một biến để tái sử dụng, tránh lặp code
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'
## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api