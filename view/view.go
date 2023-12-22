package view

import (
	"html/template"
	"strconv"
)

var Template *template.Template

func init() {
	funcMap := template.FuncMap{
		"usd": func(amount int) string {
			if amount < 0 {
				return "-$" + strconv.Itoa(-amount/100)
			}
			return "$" + strconv.Itoa(amount/100)
		},
	}

	Template = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}
