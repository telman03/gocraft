package utils

import (
	"os"
	"strings"
	"text/template"
)

// Template helper functions
var templateFuncs = template.FuncMap{
	"contains": func(slice []string, item string) bool {
		for _, s := range slice {
			if strings.EqualFold(s, item) {
				return true
			}
		}
		return false
	},
	"hasPrefix": func(s, prefix string) bool {
		return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
	},
	"hasSuffix": func(s, suffix string) bool {
		return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
	},
	"toLower": strings.ToLower,
	"toUpper": strings.ToUpper,
	"replace": strings.ReplaceAll,
}

func ApplyTemplate(srcPath, destPath string, data any) error {
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles(srcPath)
	if err != nil {
		return err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)

	// Get the base name of the template file for execution
	templateName := srcPath[strings.LastIndex(srcPath, "/")+1:]
	return tmpl.ExecuteTemplate(f, templateName, data)
}
