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
