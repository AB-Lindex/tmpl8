package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

func importObjects() ([]interface{}, error) {
	var olist []interface{}
	var err error

	for _, fn := range args.Inputs {
		var buf []byte
		if fn == "-" {
			buf, err = io.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}
			olist, err = appendEntry(olist, buf)
			if err != nil {
				return nil, err
			}
		} else {
			loads, err := load(fn, false)
			if err != nil {
				return nil, err
			}
			for _, e := range loads {
				olist, err = appendEntry(olist, []byte(e.data))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return olist, nil
}

func appendEntry(olist []interface{}, buf []byte) ([]interface{}, error) {
	var err error

	if len(buf) == 0 {
		return olist, nil
	}

	if buf[0] == '[' || buf[0] == '{' {
		var o interface{}
		err = json.Unmarshal(buf, &o)
		if err != nil {
			log.Error().Msgf("error parsing json input: %v", err)
			os.Exit(1)
		}
		if args.Split {
			if list, ok := interfaceSlice(o); ok {
				olist = append(olist, list...)
			} else {
				olist = append(olist, o)
			}
		} else {
			olist = append(olist, o)
		}
		return olist, nil
	}

	dec := yaml.NewDecoder(bytes.NewBuffer(buf))
	for {
		var o interface{}
		err = dec.Decode(&o)
		if err == nil {
			olist = append(olist, o)
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error().Msgf("error parsing yaml input: %v", err)
			os.Exit(1)
		}
	}
	return olist, nil
}
