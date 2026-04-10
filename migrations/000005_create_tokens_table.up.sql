-- Tạo bảng tokens nếu nó chưa tồn tại.
CREATE TABLE IF NOT EXISTS tokens ( 
    -- 'hash' chỉ lưu mã băm của token bằng thuật toán SHA-256 để bảo vệ dự liệu. Không bao giờ lưu trực tiếp token vào database.
    hash bytea PRIMARY KEY, 
    -- Cột 'user_id' liên kết với id của bảng 'users'. 
    -- Tính năng 'ON DELETE CASCADE' sẽ tự động xóa tất cả các token liên quan khi tài khoản người dùng đó bị xóa.
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE, 
    -- Cột 'expiry' định nghĩa thời hạn hiệu lực của token. Chúng ta sẽ giới hạn token kích hoạt trong 3 ngày.
    expiry timestamp(0) with time zone NOT NULL, 
    -- Cột 'scope' để định nghĩa loại token (ví dụ: kích hoạt hoặc xác thực)
    scope text NOT NULL
);
