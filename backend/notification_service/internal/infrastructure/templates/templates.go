package templates

import (
	"bytes"
	"html/template"
)

func Render(name string, tpl string, data interface{}) (string, error) {
	t, err := template.New(name).Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}
