package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Define một type 'envelope' để việc bọc dữ liệu trở nên tường minh hơn
type envelope map[string]any

// Viết hàm writeJSON helper.
// Hàm này nhận vào: ResponseWriter, mã HTTP status, dữ liệu cần gửi, và các Headers tùy chọn.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Sử dụng MarshalIndent để dữ liệu trả về đẹp mắt (Như đã thảo luận ở 3.2)
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Thêm ký tự xuống dòng để kết quả đẹp hơn trên terminal
	js = append(js, '\n')

	// Thêm bất kỳ header bổ sung nào khách hàng yêu cầu
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Thiết lập Header Content-Type và ghi mã Status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// 1. CHỐNG DOS (Tấn công từ chối dịch vụ)
	// Giới hạn file JSON tải lên chỉ được max 1MB. (1,048,576 bytes)
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	// 2. CHỐNG THÊM RÁC
	// Cấm khách hàng gửi những trường lạ không có trong struct
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		// Dưới đây là "bộ đồ nghề" nội soi lỗi của Go
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		// Thiếu ngoặc kép, dấu phẩy...
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		// Lỗi cấu trúc JSON bị hỏng giữa chừng
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		// Sai kiểu dữ liệu (Ví dụ gõ chữ vào trường số int)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		// Rỗng tuếch (Không gửi gì cả)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		// Đây là do cái lệnh DisallowUnknownFields() ở trên bắt được trường lạ
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// File JSON to hơn 1MB bị lệnh MaxBytesReader ngăn chặn
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		// Lỗi do bạn code sai hàm (ví dụ quên truyền con trỏ &dst)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	// 3. CHỐNG SPAM HAI LẦN
	// Tránh việc gửi kiểu: { "title": "A" } { "title": "B" }
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}
