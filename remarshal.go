// remarshal, a utility to convert between serialization formats.
// Copyright (C) 2014 Danyil Bohdan
// Adapted by Isagani Mendoza (http://itjumpstart.wordpress.com)
// License: MIT
package remarshal

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type format int

const (
	fTOML format = iota
	fYAML
	fJSON
	fPlaceholder
	fUnknown
)

// convertMapsToStringMaps recursively converts values of type
// map[interface{}]interface{} contained in item to map[string]interface{}. This
// is needed before the encoders for TOML and JSON can accept data returned by
// the YAML decoder.
func convertMapsToStringMaps(item interface{}) (res interface{}, err error) {
	switch item.(type) {
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range item.(map[interface{}]interface{}) {
			res[k.(string)], err = convertMapsToStringMaps(v)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case []interface{}:
		res := make([]interface{}, len(item.([]interface{})))
		for i, v := range item.([]interface{}) {
			res[i], err = convertMapsToStringMaps(v)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	default:
		return item, nil
	}
}

// convertNumbersToInt64 recursively walks the structures contained in item
// converting values of the type json.Number to int64 or, failing that, float64.
// This approach is meant to prevent encoders from putting numbers stored as
// json.Number in quotes or encoding large intergers in scientific notation.
func convertNumbersToInt64(item interface{}) (res interface{}, err error) {
	switch item.(type) {
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, v := range item.(map[string]interface{}) {
			res[k], err = convertNumbersToInt64(v)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case []interface{}:
		res := make([]interface{}, len(item.([]interface{})))
		for i, v := range item.([]interface{}) {
			res[i], err = convertNumbersToInt64(v)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case json.Number:
		n, err := item.(json.Number).Int64()
		if err != nil {
			f, err := item.(json.Number).Float64()
			if err != nil {
				// Can't convert to Int64.
				return item, nil
			}
			return f, nil
		}
		return n, nil
	default:
		return item, nil
	}
}

// unmarshal decodes serialized data in the format inputFormat into a structure
// of nested maps and slices.
func unmarshal(input []byte, inputFormat format) (data interface{},
	err error) {
	switch inputFormat {
	case fTOML:
		_, err = toml.Decode(string(input), &data)
	case fYAML:
		err = yaml.Unmarshal(input, &data)
		if err == nil {
			data, err = convertMapsToStringMaps(data)
		}
	case fJSON:
		decoder := json.NewDecoder(bytes.NewReader(input))
		decoder.UseNumber()
		err = decoder.Decode(&data)
		if err == nil {
			data, err = convertNumbersToInt64(data)
		}
	}
	if err != nil {
		return nil, err
	}
	return
}

// marshal encodes data stored in nested maps and slices in the format
// outputFormat.
func marshal(data interface{}, outputFormat format,
	indentJSON bool) (result []byte, err error) {
	switch outputFormat {
	case fTOML:
		buf := new(bytes.Buffer)
		err = toml.NewEncoder(buf).Encode(data)
		result = buf.Bytes()
	case fYAML:
		result, err = yaml.Marshal(&data)
	case fJSON:
		result, err = json.Marshal(&data)
		if err == nil && indentJSON {
			buf := new(bytes.Buffer)
			err = json.Indent(buf, result, "", "  ")
			result = buf.Bytes()
		}
	}
	if err != nil {
		return nil, err
	}
	return
}

//inputF and outputF can be any of the following: TOML, JSON, YAML
func Convert(input []byte, inputF, outputF string) (string, error) {

	if inputF == outputF {
		return "", errors.New("Input and output formats cannot be the same")
	}

	var inputFormat format

	switch inputF {
	case "TOML":
		inputFormat = fTOML
	case "JSON":
		inputFormat = fJSON
	case "YAML":
		inputFormat = fYAML
	default:
		inputFormat = -1
	}

	if inputFormat == -1 {
		return "", errors.New("Wrong input format: must be TOML, JSON or YAML")
	}

	var outputFormat format
	switch outputF {
	case "TOML":
		outputFormat = fTOML
	case "JSON":
		outputFormat = fJSON
	case "YAML":
		outputFormat = fYAML
	default:
		outputFormat = -1
	}

	if outputFormat == -1 {
		return "", errors.New("Wrong output format: must be TOML, JSON or YAML")
	}

	// Convert the input data from inputFormat to outputFormat.
	data, err := unmarshal(input, inputFormat)
	if err != nil {
		return "", err
	}

	indentJSON := true
	output, err := marshal(data, outputFormat, indentJSON)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
