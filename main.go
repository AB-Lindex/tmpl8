package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"text/template"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"})

	olist, err := importObjects()
	if err != nil {
		log.Error().Msgf("error reading input: %v", err)
		os.Exit(1)
	}

	var templates []entry
	for _, fn := range args.Templates {
		list, err := load(fn, true)
		if err != nil {
			log.Error().Msgf("error loading template '%s': %v", fn, err)
			os.Exit(1)
		}
		templates = append(templates, list...)
	}

	var output bytes.Buffer

	if args.Verbose {
		log.Info().Msgf("generating %d objects using %d templates", len(olist), len(templates))
	}

	t, err := prepare(templates)
	if err != nil {
		log.Error().Msgf("prepare-template-error: %v", err)
		os.Exit(1)
	}

	for _, o := range olist {
		err = process(t, o, &output)
		if err != nil {
			var execErr template.ExecError
			if errors.As(err, &execErr) {
				log.Error().Str("name", execErr.Name).Msgf("template exec-error: %v", execErr)
			} else {
				log.Error().Msgf("template-error: %v", err)
			}
			os.Exit(1)
		}
	}

	if args.Output == "" {
		_, _ = os.Stdout.Write(output.Bytes())
	} else {
		file, err := os.Create(args.Output)
		if err != nil {
			log.Error().Msgf("unable to save output to '%s': %v", args.Output, err)
			os.Exit(1)
		}
		_, _ = output.WriteTo(file)
		_ = file.Close()
	}

}

func prepare(list []entry) ([]*template.Template, error) {
	if args.VeryVerbose {
		log.Trace().Msgf("preparing %d template", len(list))
	}
	var tmpls []*template.Template

	tmpl := template.New("base").Funcs(tmpl8funcs())
	var err error

	for _, e := range list {
		if args.VeryVerbose {
			log.Trace().Msgf("parsing template '%s'...", e.name)
		}
		tmpl, err = tmpl.New(e.name).Parse(e.data)
		if err != nil {
			return nil, err
		}
		tmpls = append(tmpls, tmpl)
	}
	return tmpls, nil
}

func process(tmpls []*template.Template, o interface{}, w io.Writer) error {
	for _, tmpl := range tmpls {
		if args.VeryVerbose {
			log.Trace().Msgf("executing template '%s'...", tmpl.Name())
		}
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, o)
		if err != nil {
			return err
		}
		if !args.Raw && buf.Len() > 0 {
			bytes := buf.Bytes()
			if bytes[len(bytes)-1] != '\n' {
				buf.WriteByte('\n')
			}
		}
		_, _ = buf.WriteTo(w)
	}
	return nil
}

func interfaceSlice(slice interface{}) ([]interface{}, bool) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, false
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil, false
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, true
}
