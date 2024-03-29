package types

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"text/template"
)

type TemplateFile string

func (t TemplateFile) parse(values map[string]interface{}) (ret []byte, err error) {
	if ret, err = os.ReadFile(string(t)); err != nil {
		return
	}

	var tmpl *template.Template
	if tmpl, err = template.New("unit").Parse(string(ret)); err != nil {
		return
	}

	buf := bytes.Buffer{}
	if err = tmpl.Execute(&buf, values); err != nil {
		return
	}
	if ret, err = io.ReadAll(&buf); err != nil {
		return
	}
	return
}

type TemplateParseObjectFunc func(data []byte) error

func (t TemplateFile) Parse(values map[string]interface{}) (ret map[string]interface{}, err error) {
	var b []byte
	if b, err = t.parse(values); err != nil {
		return
	}
	err = json.Unmarshal(b, &ret)
	return
}

func (t TemplateFile) ParseString(values map[string]interface{}) (ret string, err error) {
	var b []byte
	if b, err = t.parse(values); err != nil {
		return
	}
	ret = string(b)
	return
}

func (t TemplateFile) ParseObject(values map[string]interface{}, parse TemplateParseObjectFunc) (err error) {
	var b []byte
	if b, err = t.parse(values); err != nil {
		return
	}
	err = parse(b)
	return
}
