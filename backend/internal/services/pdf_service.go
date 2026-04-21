package services

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

func ReadPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("error al abrir pdf: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("error al leer texto plano: %w", err)
	}

	_, err = buf.ReadFrom(b)
	return buf.String(), err
}
