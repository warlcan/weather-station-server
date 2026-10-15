package templates

import (
	"log"
	"text/template"
)

func GetTemplate(templatePath string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Error parsing template file: %v", err)
		return nil, err
	}
	return tmpl, nil
}
