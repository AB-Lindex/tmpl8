package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

type arguments struct {
	Inputs      []string `arg:"-i,--input,separate" help:"filename of objects to import and process (use '-' for stdin)"`
	Templates   []string `arg:"required,positional" help:"filename.tmpl | @filenames.lst | k8s:namespace/configmap" placeholder:"TEMPLATE"`
	Output      string   `arg:"-o,--output" help:"destination filename" placeholder:"FILE"`
	Verbose     bool     `arg:"-v,--verbose" help:"verbose output"`
	VeryVerbose bool     `arg:"-t,--trace" help:"trace output"`
	Raw         bool     `arg:"-r,--raw" help:"will NOT ensure newline and end of each block"`
	Split       bool     `arg:"-s,--split" help:"split json-array into separate 'documents'"`
	NoInput     bool     `arg:"-z" help:"no input (equals -i '?{}')"`
}

func (arguments) Description() string {
	return `tmpl8 - Generic (and Kubernetes-friendly) Templating Engine using the go text/template and Sprig functions

Supported input formats: JSON and YAML
`
}

var args arguments

func init() {
	arg.MustParse(&args)

	if len(args.Templates) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: tmpl8 <file> [<file>...]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "   <file> is either a real file,")
		fmt.Fprintln(os.Stderr, "   or '@filename' where 'filename' contains list or files to process")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  The input object is read from <stdin>")
		os.Exit(1)
	}

	if args.NoInput {
		args.Inputs = append(args.Inputs, "?{}")
	}

	if args.VeryVerbose {
		args.Verbose = true
	}

	if len(args.Inputs) == 0 {
		args.Inputs = append(args.Inputs, "-")
	}
}
