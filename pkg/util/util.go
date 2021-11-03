package util

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func RenderTemplate(base string, data interface{}) (string, error) {
	tmpl, err := template.New("command").Funcs(sprig.TxtFuncMap()).Parse(base)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, data)
	return buf.String(), err
}
