# Project Rules - Greenlight API

Mọi quy tắc trong file này phải được AI tuân thủ tuyệt đối trong suốt quá trình phát triển dự án.

## 1. Nguồn dữ liệu và Code (Source of Truth)
- **BẮT BUỘC** lấy nguồn từ file `book.txt` (được trích xuất từ sách "Let's Go Further" của Alex Edwards). Đọc hết tất cả thông tin, nhiều quá thì chia nhỏ ra để đọc và phản hồi, chứ không được bỏ qua bất kì phần nào. Phải trích xuất code đi kèm. Các đầu mục phản hồi phải được chia theo đúng đầu mục trong sách. Cuối mỗi phản hồi phải đặt câu hỏi để đảm bảo người dùng hiểu mục đích của chương. không tự ý sửa code.
- Không tự ý sử dụng các thư viện ngoài hoặc logic khác nếu sách không hướng dẫn (trừ khi có yêu cầu riêng từ người dùng).
- Mọi giải pháp code phải được đối chiếu với chương/mục tương ứng trong sách.
- Yêu cầu đọc chương nào thì chỉ đọc chương đấy không được phép đọc qua chương khác.
- Không bao giờ tự ý viết code vào file trừ khi có yêu cầu.



## 2. Phong cách lập trình (Coding Style)
- Tuân thủ các patterns về Dependency Injection (thông qua struct `application`).
- Sử dụng các helper được định nghĩa trong `helpers.go` và `errors.go`.
- Luôn ưu tiên xử lý lỗi (error handling) một cách chi tiết và trả về JSON.

---
*(Bạn có thể thêm các rules tiếp theo vào dưới đây)*
