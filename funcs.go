package main

import (
	"os"

	sprig "github.com/Masterminds/sprig/v3"
)

var myFuncs = map[string]interface{}{
	"version":  versionFunc,
	"readfile": readFile,
	"isfunc":   isFunc,
}

var allFuncs map[string]interface{}

func tmpl8funcs() map[string]interface{} {
	allFuncs = sprig.TxtFuncMap()
	for k, v := range myFuncs {
		allFuncs[k] = v
	}
	return allFuncs
}

func isFunc(a string) bool {
	_, ok := allFuncs[a]
	return ok
}

func readFile(name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(data)
}
