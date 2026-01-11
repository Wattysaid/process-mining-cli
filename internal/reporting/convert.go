package reporting

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// MarkdownToHTML converts a markdown file to a standalone HTML file.
func MarkdownToHTML(markdownPath string, htmlPath string, title string) error {
	if markdownPath == "" || htmlPath == "" {
		return errors.New("markdown and html paths are required")
	}
	input, err := os.ReadFile(markdownPath)
	if err != nil {
		return err
	}
	body := blackfriday.Run(input)
	pageTitle := title
	if pageTitle == "" {
		pageTitle = strings.TrimSuffix(filepath.Base(markdownPath), filepath.Ext(markdownPath))
	}
	var buf bytes.Buffer
	buf.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	buf.WriteString(fmt.Sprintf("<title>%s</title>\n", pageTitle))
	buf.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	buf.WriteString("<style>body{font-family:Arial,Helvetica,sans-serif;max-width:960px;margin:40px auto;padding:0 16px;line-height:1.6;}h1,h2,h3{margin-top:1.4em;}table{border-collapse:collapse;width:100%;}th,td{border:1px solid #ccc;padding:6px 8px;text-align:left;}code{background:#f4f4f4;padding:2px 4px;border-radius:4px;}</style>\n")
	buf.WriteString("</head>\n<body>\n")
	buf.Write(body)
	buf.WriteString("\n</body>\n</html>\n")
	if err := os.WriteFile(htmlPath, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return nil
}

// MarkdownToPDF converts a markdown file to PDF using pandoc if available.
func MarkdownToPDF(markdownPath string, pdfPath string) error {
	if markdownPath == "" || pdfPath == "" {
		return errors.New("markdown and pdf paths are required")
	}
	if _, err := exec.LookPath("pandoc"); err != nil {
		return fmt.Errorf("pandoc not found in PATH")
	}
	cmd := exec.Command("pandoc", markdownPath, "-o", pdfPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
