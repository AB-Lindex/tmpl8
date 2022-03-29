package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/rs/zerolog/log"
)

func execCmd(name string, args ...string) (stdout []byte, stderr []byte, err error) {
	cmd := exec.Command(name, args...)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	var wg sync.WaitGroup

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	var outOk bool
	var errOk bool

	wg.Add(2)
	go execPipe(outPipe, &bufOut, &outOk, &wg)
	go execPipe(errPipe, &bufErr, &errOk, &wg)

	err = cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start 'kubectl': %v", err)
	}

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		return bufOut.Bytes(), bufErr.Bytes(), fmt.Errorf("'kubectl' failed: %v", err)
	}

	return bufOut.Bytes(), bufErr.Bytes(), nil
}

func execPipe(pipe io.Reader, buffer *bytes.Buffer, ok *bool, wg *sync.WaitGroup) {
	defer wg.Done()
	b := make([]byte, 1024)
	for {
		n, err := pipe.Read(b)
		buffer.Write(b[:n])
		if err != nil {
			*ok = err == io.EOF
			if !*ok {
				log.Warn().Msgf("execPipe error: %v", err)
			}
			return
		}
	}
}
