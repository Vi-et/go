package validator

import (
	"regexp"
)

// (Tạm lưu lại một Regex chuẩn chỉnh của HTML5 để dùng cho xác thực Email sau này)
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Khởi tạo struct chứa từ điển lưu trữ các lỗi validation.
type Validator struct {
	Errors map[string]string
}

// Dùng hàm New() để cấp phát bộ nhớ cho map trống.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Trả về true nếu map đang rỗng (tức là không chặn được bất kỳ lỗi nào).
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Thêm 1 lỗi vào bản đồ, nhưng chỉ thêm nếu trường key đó chưa có lỗi.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Hàm Check() là cánh tay phải của chúng ta. Nếu 'ok' là false, nó lập tức AddError.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Trả về true nếu Chuỗi chỉ định nằm trong mảng các Chuỗi cho phép.
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Trả về true nếu chuỗi khớp với khuôn mẫu Regex
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Trả về true nếu tất cả các phần tử trong mảng slice đều là duy nhất (không trùng nhau).
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
