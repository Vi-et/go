package data

import (
	"errors"
	"strconv"
	"strings"
)

// Định nghĩa cái lỗi riêng nếu chữ gửi tới sai fomat.
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// ... (code cũ)

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// BƯỚC 1: Dữ liệu JSON luôn ngậm ở 2 đầu là ngoặc kép (vd: '"107 mins"').
	// Chúng ta phải gỡ lớp ngoặc kép đi bằng hàm strconv.Unquote
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// BƯỚC 2: Cắt làm đôi dựa trên dấu cách. Ta sẽ có 2 phần: "107" và "mins".
	parts := strings.Split(unquotedJSONValue, " ")

	// BƯỚC 3: Rào lỗi. Nếu bị cắt ra lớn/nhỏ hơn 2 phần, hoặc chữ cuối không phải chữ "mins" thì báo lỗi ngay.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// BƯỚC 4: Ép phần số "107" từ dạng chữ sang số nguyên (int32)
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// BƯỚC 5: Gán giá trị r bằng trị số int32 vừa ép thành công.
	// Dấu sao * ở đây nghĩa là giải tham chiếu trỏ tới giá trị nền của con trỏ The pointer.
	*r = Runtime(i)

	return nil
}
