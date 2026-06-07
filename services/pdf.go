package services

import (
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractResumeText(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(textReader)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}
