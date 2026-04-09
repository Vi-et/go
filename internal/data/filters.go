package data

import (
	"math"
	"strings"

	"greenlight.example.com/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string // Thêm danh sách an toàn để chứa các giá trị sort hợp lệ
}

// ValidateFilters là hàm kiểm tra tính hợp lệ của các bộ lọc tìm kiếm
func ValidateFilters(v *validator.Validator, f Filters) {
	// Kiểm tra page và page_size có nằm trong khoảng hợp lý không
	v.Check(f.Page > 0, "page", "phải lớn hơn không")
	v.Check(f.Page <= 10_000_000, "page", "tối đa là 10 triệu")
	v.Check(f.PageSize > 0, "page_size", "phải lớn hơn không")
	v.Check(f.PageSize <= 100, "page_size", "tối đa là 100")
	// Sử dụng helper validator.In() để kiểm tra xem giá trị f.Sort
	// có nằm trong danh sách an toàn (Safelist) hay không.
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "giá trị sắp xếp không hợp lệ")
}

// Hàm trả về tên cột sắp xếp (bỏ dấu "-" nếu có)
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	// Chốt chặn cuối bảo vệ khỏi SQL Injection
	panic("unsafe sort parameter: " + f.Sort)
}

// Hàm trả về chiều sắp xếp: "ASC" hoặc "DESC"
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// Hàm trả về LIMIT (số lượng bản ghi tối đa mỗi trang)
func (f Filters) limit() int {
	return f.PageSize
}

// Hàm trả về OFFSET (số bản ghi cần bỏ qua để đến đúng trang)
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Struct chứa thông tin phân trang trả về cho client
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// Hàm tính toán metadata phân trang dựa trên tổng số bản ghi, trang hiện tại và kích thước trang
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{} // Trả về struct rỗng nếu không có bản ghi nào
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
