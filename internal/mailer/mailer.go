package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// Sử dụng //go:embed để đưa thư mục "templates" vào hệ thống tệp nhúng.
// LƯU Ý: Lệnh chú thích này phải nằm NGAY TRÊN dòng khai báo biến.
//
//go:embed "templates"
var templateFS embed.FS

// Định nghĩa struct Mailer chứa bộ kết nối dialer (cổng vào SMTP) và thông tin người gửi.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// Khởi tạo Mailer mới
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second // Giới hạn connection gửi mail mất tối đa 5 giây.

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Phương thức Send để xuất bản một email
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// Trích xuất file template mục tiêu từ file system nhúng "templateFS"
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Tiến hành nạp dữ liệu động (Dynamic Data) vào block template "subject".
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Nạp dữ liệu động vào block "plainBody"
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Nạp dữ liệu động vào block "htmlBody"
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Thiết lập cấu trúc bức Mail
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	// AddAlternative LUÔN phải đặt sau SetBody.
	msg.AddAlternative("text/html", htmlBody.String())

	// Gọi phương thức DialAndSend để liên lạc mở kết nối SMTP và đẩy gói tin.
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
