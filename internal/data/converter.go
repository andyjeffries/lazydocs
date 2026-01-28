package data

import (
	"strings"

	htmltomd "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// Converter handles HTML to Markdown conversion
type Converter struct{}

// NewConverter creates a new HTML to Markdown converter
func NewConverter() *Converter {
	return &Converter{}
}

// Convert converts HTML content to Markdown
func (c *Converter) Convert(html string) (string, error) {
	// Handle empty content
	if strings.TrimSpace(html) == "" {
		return "", nil
	}

	md, err := htmltomd.ConvertString(html)
	if err != nil {
		return "", err
	}

	// Clean up the markdown
	md = cleanMarkdown(md)

	return md, nil
}

// cleanMarkdown performs post-processing on the converted markdown
func cleanMarkdown(md string) string {
	// Remove excessive blank lines
	lines := strings.Split(md, "\n")
	var result []string
	blankCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				result = append(result, line)
			}
		} else {
			blankCount = 0
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}
