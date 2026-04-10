# Project Rules - Greenlight API

Mọi quy tắc trong file này phải được AI tuân thủ tuyệt đối trong suốt quá trình phát triển dự án.

## 1. Nguồn dữ liệu và Code (Source of Truth)
- **BẮT BUỘC** lấy nguồn từ file `book.txt` (được trích xuất từ sách "Let's Go Further" của Alex Edwards). Đọc hết tất cả thông tin, nhiều quá thì chia nhỏ ra để đọc và phản hồi, chứ không được bỏ qua bất kì phần nào. Phải trích xuất code đi kèm. Các đầu mục phản hồi phải được chia theo đúng đầu mục trong sách. Cuối mỗi phản hồi phải đặt câu hỏi để đảm bảo người dùng hiểu mục đích của chương và đừng bao giờ hỏi câu hỏi có/không hoặc đã hiểu chưa hãy đặt một câu hỏi dựa vào nội dung trong chương. comment code phải bằng tiếng việt. Phải trích xuất đầy đủ các thông tin vì sao lại làm như trong sách nếu tác giả có đề cập và giải thích cho người đọc hiểu với giả định rằng người đọc không hiểu gì.
- Không tự ý sử dụng các thư viện ngoài hoặc logic khác nếu sách không hướng dẫn (trừ khi có yêu cầu riêng từ người dùng).
- Mọi giải pháp code phải được đối chiếu với chương/mục tương ứng trong sách.
- Yêu cầu đọc chương nào thì chỉ đọc chương đấy không được phép đọc qua chương khác.

## ⛔ TUYỆT ĐỐI KHÔNG TỰ Ý VIẾT / SỬA CODE
- **KHÔNG BAO GIỜ** tự ý tạo file, sửa file, hoặc viết code vào bất kỳ file nào trong project trừ khi có yêu cầu trực tiếp.
- Khi hướng dẫn một chapter, **CHỈ ĐƯỢC PHÉP** hiển thị code dưới dạng code block trong chat để người dùng tự tay thực hiện.
- **KHÔNG** dùng các tool: `write_to_file`, `replace_file_content`, `multi_replace_file_content`, `run_command` để thay đổi source code.
- Chỉ được dùng `view_file`, `list_dir`, `grep_search` để đọc và hiểu codebase hiện tại.
- Vi phạm quy tắc này là vi phạm nghiêm trọng nhất.

## 2. Phong cách lập trình (Coding Style)
- Tuân thủ các patterns về Dependency Injection (thông qua struct `application`).
- Sử dụng các helper được định nghĩa trong `helpers.go` và `errors.go`.
- Luôn ưu tiên xử lý lỗi (error handling) một cách chi tiết và trả về JSON.

## 3. Tóm tắt lý thuyết sau mỗi chương (Chapter Summary)
- **BẮT BUỘC**: Sau khi hướng dẫn xong một chương, phải tạo file tóm tắt lý thuyết của chương đó.
- File được lưu tại thư mục `summaries/` với tên theo format: `chapter_<số chương>_<tên chương>.txt` (ví dụ: `summaries/chapter_12_1_rate_limiting.txt`).
- Nội dung file **CHỈ bao gồm lý thuyết**: các khái niệm, giải thích, lý do tại sao, cách hoạt động — **KHÔNG chứa bất kỳ đoạn code nào**.
- **ĐẶC BIỆT**: Khuyến khích mô tả cách thức hoạt động bằng sơ đồ luồng (flow) dạng hình vẽ (sử dụng ASCII art trực quan) trong file để dễ lập luận và hình dung.
- Nội dung phải được viết bằng tiếng Việt, rõ ràng, dễ hiểu như một bản ghi chú học tập.
- Đây là trường hợp **NGOẠI LỆ DUY NHẤT** được phép dùng `write_to_file` — chỉ để tạo file tóm tắt trong thư mục `summaries/`, không được dùng cho bất kỳ mục đích nào khác.

---
*(Bạn có thể thêm các rules tiếp theo vào dưới đây)*
