package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type entry struct {
	name string
	data string
}

func load(fn string, allowK8s bool) ([]entry, error) {
	if len(fn) == 0 {
		return nil, nil
	}

	var result []entry

	if strings.HasPrefix(fn, "?") {
		result = append(result, entry{"{inline}", fn[1:]})
		return result, nil
	}

	// @filename -> read filename, each line is a new filename to add
	if strings.HasPrefix(fn, "@") {
		return loadFiles(fn[1:], allowK8s)
	}

	if allowK8s {
		if strings.HasPrefix(fn, "k8s:") {
			return loadK8s(fn[4:])
		}
	}

	// regular file
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	result = append(result, entry{fn, string(buf)})

	return result, nil
}

func loadFiles(fn string, allowK8s bool) ([]entry, error) {
	var result []entry

	if args.Verbose {
		log.Info().Msgf("expanding '%s'...", fn)
	}

	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		inner, err := load(scanner.Text(), allowK8s)
		if err != nil {
			return nil, err
		}
		result = append(result, inner...)
	}
	return result, nil
}
