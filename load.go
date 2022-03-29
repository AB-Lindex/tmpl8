package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
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

func loadK8s(fn string) ([]entry, error) {
	if args.Verbose {
		log.Info().Msgf("loading k8s-config '%s'...", fn)
	}

	parts := strings.Split(fn, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("k8s-format is 'k8s:namespace/configname', not 'k8s:%s'", fn)
	}

	stdout, stderr, err := execCmd("kubectl", "get", "configmap", "-n", parts[0], parts[1], "-o", "json")
	// stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.WithLevel(zerolog.PanicLevel).Msgf("kubectl-error: %s", strings.TrimSuffix(string(stderr), "\n"))
		return nil, err
	}

	var cm configmap
	err = json.Unmarshal(stdout, &cm)
	if err != nil {
		return nil, err
	}

	var result []entry

	for name, data := range cm.Data {
		result = append(result, entry{fmt.Sprintf("k8s:%s/%s/%s", parts[0], parts[1], name), data})
	}

	return result, nil
}

type configmap struct {
	Data map[string]string `json:"data"`
}
