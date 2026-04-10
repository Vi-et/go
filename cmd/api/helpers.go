package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.example.com/internal/data"
	"greenlight.example.com/internal/validator"
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

// Hàm trợ giúp readString() trả về giá trị kiểu chuỗi từ query string,
// hoặc giá trị mặc định (defaultValue) nếu tham số không tồn tại.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Lấy giá trị của tham số từ query string.
	s := qs.Get(key)
	// Nếu giá trị trả về rỗng, chứng tỏ client không cung cấp, ta trả về mặc định.
	if s == "" {
		return defaultValue
	}
	return s
}

// Hàm trợ giúp readCSV() đọc giá trị phân tách bằng dấu phẩy và tách nó thành danh sách các chuỗi (slice).
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	// Tách chuỗi theo dấu phẩy và trả về dưới dạng mảng (slice).
	return strings.Split(csv, ",")
}

// Hàm trợ giúp readInt() lấy giá trị chuỗi từ query string và ép kiểu sang số nguyên int.
// Đồng thời, nếu không phải số hợp lệ, nó sẽ lưu lỗi đó vào thẳng biến Validator truyền vào (v).
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	// Ép kiểu chuỗi sang kiểu integer.
	i, err := strconv.Atoi(s)
	if err != nil {
		// Thêm thông báo lỗi vào validator instace nếu quá trình ép kiểu thất bại.
		v.AddError(key, "phải là số nguyên/integer")
		return defaultValue
	}
	return i
}

// The background() helper chấp nhận bất kỳ hàm ẩn danh (anonymous function) nào làm tham số.
func (app *application) background(fn func()) {

	// 1. Khai báo có 1 Goroutine chuẩn bị được tung vào nền
	app.wg.Add(1)
	// Khởi tạo một background goroutine.
	go func() {
		defer app.wg.Done()
		// Triển khai Defer để khôi phục mọi panic có thể xảy ra và ghi chép lại lỗi,
		// ngăn không cho làm sập ứng dụng.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Thực thi hàm truyền vào (Thực chất là tiến trình Send Email sẽ nằm ở đây).
		fn()
	}()
}

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
