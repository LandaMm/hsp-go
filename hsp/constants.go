package hsp

import (
	"errors"
	"fmt"
	"strings"
)

// DO NOT CHANGE THIS
const HSP_PORT = "998"

const (
	H_STATUS      = "status"
	H_DATA_FORMAT = "data-format"
	H_ROUTE       = "route"
)

const (
	DF_BYTES = "bytes"
	DF_TEXT  = "text"
	DF_JSON  = "json"
)

const (
	E_UTF8 = "utf-8"
)

const (
	STATUS_SUCCESS     = 0
	STATUS_NOTFOUND    = 69
	STATUS_INTERNALERR = 129
)

var DATA_FORMATS map[string]string = map[string]string{
	"bytes": DF_BYTES,
	"text":  DF_TEXT,
	"json":  DF_JSON,
}

var ENCODINGS map[string]string = map[string]string{
	"utf-8": E_UTF8,
}

type DataFormat struct {
	Format   string
	Encoding string
}

func TextDataFormat() *DataFormat {
	return &DataFormat{Format: DF_TEXT, Encoding: E_UTF8}
}

func JsonDataFormat() *DataFormat {
	return &DataFormat{Format: DF_JSON, Encoding: E_UTF8}
}

func BytesDataFormat() *DataFormat {
	return &DataFormat{Format: DF_BYTES}
}

func ParseDataFormat(format string) (*DataFormat, error) {
	parts := strings.Split(format, ":")
	if len(parts) != 2 {
		if format == "bytes" {
			return &DataFormat{
				Format: DF_BYTES,
			}, nil
		}
		return nil, errors.New("Invalid data format header")
	}

	f, ok := DATA_FORMATS[parts[0]]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown data format: %s", parts[0]))
	}

	encoding, ok := ENCODINGS[parts[1]]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown payload encoding: %s", parts[1]))
	}

	return &DataFormat{
		Format:   f,
		Encoding: encoding,
	}, nil
}

func (df *DataFormat) String() string {
	if df.Format == DF_BYTES {
		return df.Format
	}
	return fmt.Sprintf("%s:%s", df.Format, df.Encoding)
}
