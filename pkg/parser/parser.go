package parser

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/davidcassany/updateinfo-parser/pkg/types"
)

const updateToken = "update"
const defaultTmpl = `--------------------------------------------------------------------------------
{{.Title}}

ID: {{.ID}}
Type: {{.Type}}
Severity: {{.Severity}}
Date: {{.Issued.Date}}

Description:
{{.Description}}

{{if .References}}Issues:{{range .References}}
  * {{.Type}}: [{{.ID}}] {{.Title}}{{end}}

{{end}}`

type Config struct {
	beforeDate   time.Time
	afterDate    time.Time
	pkgWhiteList []string
	output       io.Writer
	template     *template.Template
	updateXML    string
	updateType   string
}

const DateLayout = "2006-01-02"

func NewConfig(updateXML, beforeStr, afterStr, packagesFile, tmpl, output, updateType string) (*Config, error) {
	cfg := Config{}

	if _, err := os.Stat(updateXML); err != nil {
		return nil, fmt.Errorf("could not fild updateinfo file '%s'", updateXML)
	}
	cfg.updateXML = updateXML

	before, err := time.Parse(DateLayout, beforeStr)
	if err != nil {
		return nil, fmt.Errorf("failed parsing before date '%s': %v", beforeStr, err)
	}
	cfg.beforeDate = before

	after, err := time.Parse(DateLayout, afterStr)
	if err != nil {
		return nil, fmt.Errorf("failed parsing after date '%s': %v", afterStr, err)
	}
	cfg.afterDate = after

	packages, err := ReadPackagesFile(packagesFile)
	if err != nil {
		return nil, fmt.Errorf("failed reading packages file '%s': %v", packagesFile, err)
	}
	cfg.pkgWhiteList = packages

	var t *template.Template
	if tmpl != "" {
		t, err = template.ParseFiles(tmpl)
		if err != nil {
			return nil, fmt.Errorf("failed parsing template file '%s': %v", tmpl, err)
		}
	} else {
		t, _ = template.New("update").Parse(defaultTmpl)
	}
	cfg.template = t

	if output == "" {
		cfg.output = os.Stdout
	}

	cfg.updateType = updateType

	return &cfg, nil
}

func ReadPackagesFile(pkgFile string) ([]string, error) {
	packages := []string{}

	if pkgFile == "" {
		return packages, nil
	}

	file, err := os.Open(pkgFile)
	if err != nil {
		return nil, fmt.Errorf("failed opening file '%s': %v", pkgFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		packages = append(packages, strings.TrimSpace(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning lines on file '%s': %v", pkgFile, err)
	}
	return packages, nil
}

func Parse(cfg *Config) error {
	reader, err := os.Open(cfg.updateXML)
	if err != nil {
		return err
	}
	defer reader.Close()

	d := xml.NewDecoder(reader)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return fmt.Errorf("decoding token: %v", tokenErr)
		}
		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == updateToken {
				u := types.Update{}
				if err := d.DecodeElement(&u, &t); err != nil {
					return fmt.Errorf("decoding element %q: %v", t.Name.Local, err)
				}
				if cfg.updateType != "" && u.Type != cfg.updateType {
					continue
				}
				if u.Issued.Date == nil {
					continue
				}
				uDate := time.Time(*u.Issued.Date)
				if uDate.Before(cfg.beforeDate) && uDate.After(cfg.afterDate) {
					var pkgMatch bool
					for _, pkg := range u.Packages {
						if slices.Contains(cfg.pkgWhiteList, pkg.Name) {
							pkgMatch = true
							break
						}
					}

					if len(cfg.pkgWhiteList) == 0 || pkgMatch {
						err := cfg.template.Execute(cfg.output, &u)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
