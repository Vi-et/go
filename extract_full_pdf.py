import sys
from pypdf import PdfReader

# Đọc toàn bộ PDF
pdf_path = "Let's Go Further -- Alex Edwards -- 0474ed768dd5a4100eab736f02614bc6 -- Anna’s Archive.pdf"
reader = PdfReader(pdf_path)

# Mở file book.txt và tuần tự ghi từng trang vào (để tiết kiệm RAM)
with open("book.txt", "w", encoding="utf-8") as f:
    for i in range(len(reader.pages)):
        text = reader.pages[i].extract_text()
        if text:
            f.write(text + "\n")
