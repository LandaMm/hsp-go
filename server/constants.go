
package server

const (
	H_STATUS = "status"
	H_DATA_FORMAT = "data-format"
	H_ROUTE = "route"
)

const (
	DF_BYTES = "bytes"
	DF_TEXT = "text"
	DF_JSON = "json"
)

const (
	E_UTF8 = "utf-8"
)

const (
	STATUS_SUCCESS = 0
	STATUS_NOTFOUND = 69
	STATUS_INTERNALERR = 129
)

var DATA_FORMATS map[string]string = map[string]string{
	"bytes": DF_BYTES,
	"text": DF_TEXT,
	"json": DF_JSON,
}

var ENCODINGS map[string]string = map[string]string{
	"utf-8": E_UTF8,
}

type DataFormat struct {
	Format string
	Encoding string
}

