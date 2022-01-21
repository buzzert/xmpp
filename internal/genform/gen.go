// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// The genform command creates form fields from a registry.
package main // import "mellium.im/xmpp/internal/genform"

import (
	"bytes"
	"encoding/xml"
	"flag"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const tmpl = `// Code generated by "genform{{if gt (len .Args) 0}} {{end}}{{.Args}}"; DO NOT EDIT.

package {{.Pkg}}

import (
	"mellium.im/xmpp/form"
)

{{- range .Forms }}
{{ $func := (index $.Vars .Name) }}
	// {{$func}} creates a new configuration form.
	// Desc: {{.Desc | trimspace | comment }}.
	func {{$func}}() *form.Data {
return	form.New(
	{{- range .Field }}
		form.
		{{- if eq .Type "boolean" -}}
		Boolean({{printf "%q" .Var}},
		{{- else if eq .Type "fixed" -}}
		Fixed(
		{{- else if eq .Type "hidden" -}}
		Hidden({{printf "%q" .Var}},
		{{- else if eq .Type "jid-multi" -}}
		JIDMulti({{printf "%q" .Var}},
		{{- else if eq .Type "jid-single" -}}
		JID({{printf "%q" .Var}},
		{{- else if eq .Type "list-multi" -}}
		ListMulti({{printf "%q" .Var}},
		{{- else if eq .Type "list-single" -}}
		List({{printf "%q" .Var}},
		{{- else if eq .Type "text-multi" -}}
		TextMulti({{printf "%q" .Var}},
		{{- else if eq .Type "text-private" -}}
		TextPrivate({{printf "%q" .Var}},
		{{- else if eq .Type "text-single" -}}
		Text({{printf "%q" .Var}},
		{{- end }}
		form.Label({{printf "%q" .Var}}),
		{{- if .Label }}
		form.Desc({{printf "%q" (unwrap .Label)}}),{{ end }}
		{{- range .Option }}
		form.ListItem({{printf "%q" (unwrap .Label)}}, {{printf "%q" .Value}}),
		{{- end }}
		),
	{{- end }}
	)
	}
{{ end }}`

type form struct {
	XMLName xml.Name `xml:"form_type"`
	Name    string   `xml:"name"`
	Doc     string   `xml:"doc"`
	Desc    string   `xml:"desc"`
	Field   []struct {
		Var    string `xml:"var,attr"`
		Type   string `xml:"type,attr"`
		Label  string `xml:"label,attr"`
		Option []struct {
			XMLName xml.Name `xml:"option"`
			Label   string   `xml:"label,attr"`
			Value   string   `xml:"value"`
		} `xml:"option"`
	} `xml:"field"`
}

type registry struct {
	XMLName xml.Name `xml:"registry"`
	Form    []form   `xml:"form_type"`
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("genform: ")
	var (
		outFile = "formfields.go"
		names   = ""
		prefix  = ""
		// We should be using the generated registry at
		// https://xmpp.org/registrar/formtypes.xml, but it has a bug.
		// See: https://github.com/xsf/registrar/pull/40
		regURL = `https://raw.githubusercontent.com/xsf/registrar/master/formtypes.xml`
		tmpDir = os.TempDir()
		noFmt  bool
	)

	flag.StringVar(&outFile, "filename", outFile, "filename to generate")
	flag.StringVar(&tmpDir, "tmp", tmpDir, "A temporary directory to downlaod files to")
	flag.StringVar(&names, "names", names, "comma separated list of Name:Var pairs")
	flag.StringVar(&prefix, "prefix", prefix, "prefix to prepend to all names")
	flag.StringVar(&regURL, "formtypes", regURL, "A link to the formtypes registry")
	flag.BoolVar(&noFmt, "nofmt", noFmt, "Disables code formatting")
	flag.Parse()

	fd, err := openOrDownload(regURL, tmpDir)
	if err != nil {
		log.Fatalf("error downloading (or opening) %s in %s: %v", regURL, tmpDir, err)
	}
	/* #nosec */
	defer fd.Close()

	reg := registry{}
	d := xml.NewDecoder(fd)
	d.Strict = false
	var start xml.StartElement
	for {
		var ok bool
		tok, err := d.Token()
		if err != nil {
			log.Fatalf("error popping registry tokens: %v", err)
		}
		start, ok = tok.(xml.StartElement)
		if ok && start.Name.Local == "registry" {
			break
		}
	}
	if err = d.DecodeElement(&reg, &start); err != nil {
		log.Fatalf("error decoding registry: %v", err)
	}

	formNames := make(map[string]string)
	for _, pair := range strings.Split(names, ",") {
		idx := strings.LastIndexByte(pair, ':')
		if idx == -1 {
			log.Printf("skipping invalid pair %q", pair)
			continue
		}
		formNames[prefix+pair[:idx]] = pair[idx+1:]
	}
	if len(formNames) == 0 {
		log.Fatal("no forms specified")
	}

	// Filter the form names we've parsed to only the ones specified in the
	// argument (if any).
	n := 0
	for _, f := range reg.Form {
		if _, ok := formNames[f.Name]; ok {
			reg.Form[n] = f
			n++
		}
	}
	reg.Form = reg.Form[:n]

	pkgs, err := packages.Load(nil, ".")
	if err != nil {
		log.Fatalf("error loading package: %v", err)
	}
	pkg := pkgs[0]

	parsedTmpl, err := template.New("out").Funcs(map[string]interface{}{
		"trimspace": strings.TrimSpace,
		"unwrap": func(s string) string {
			return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
		},
		"comment": func(s string) string {
			fields := strings.Split(s, "\n")
			for i, f := range fields {
				if i == 0 {
					continue
				}

				fields[i] = "//       " + strings.TrimSpace(f)
			}
			return strings.Join(fields, "\n")
		},
	}).Parse(tmpl)
	if err != nil {
		log.Fatalf("error parsing template: %v", err)
	}

	var buf bytes.Buffer
	fd, err = os.Create(outFile)
	if err != nil {
		log.Fatalf("error creating file %q: %v", outFile, err)
	}
	err = parsedTmpl.Execute(&buf, struct {
		Args  string
		Pkg   string
		Forms []form
		Vars  map[string]string
	}{
		Args:  strings.Join(os.Args[1:], " "),
		Pkg:   pkg.Name,
		Forms: reg.Form,
		Vars:  formNames,
	})
	if err != nil {
		log.Fatalf("error executing template: %v", err)
	}

	if noFmt {
		_, err = io.Copy(fd, &buf)
		if err != nil {
			log.Fatalf("error writing file: %v", err)
		}
	} else {
		fmtBuf, err := format.Source(buf.Bytes())
		if err != nil {
			log.Fatalf("error formatting source: %v", err)
		}

		_, err = io.Copy(fd, bytes.NewReader(fmtBuf))
		if err != nil {
			log.Fatalf("error writing formatted file: %v", err)
		}
	}
}

// opens the provided registry URL (downloading it if it doesn't exist).
func openOrDownload(catURL, tmpDir string) (*os.File, error) {
	registryXML := filepath.Join(tmpDir, filepath.Base(catURL))
	/* #nosec */
	fd, err := os.Open(registryXML)
	if err != nil {
		/* #nosec */
		fd, err = os.Create(registryXML)
		if err != nil {
			return nil, err
		}
		// If we couldn't open it for reading, attempt to download it.

		/* #nosec */
		resp, err := http.Get(catURL)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(fd, resp.Body)
		if err != nil {
			return nil, err
		}
		/* #nosec */
		resp.Body.Close()
		_, err = fd.Seek(0, 0)
		if err != nil {
			return nil, err
		}
	}
	return fd, err
}
