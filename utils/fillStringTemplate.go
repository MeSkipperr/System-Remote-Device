package utils
import "strings"

func FillTemplate(template string, values map[string]string) string {
	for key, val := range values {
		template = strings.ReplaceAll(template, "{"+key+"}", val)
	}
	return template
}
