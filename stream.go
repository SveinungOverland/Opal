package opal

import (
	"opal/frame"
	"opal/hpack"
	"opal/http"
	"strings"
)

type StreamState uint8

const (
	idle StreamState = iota + 1
	reservedLocal
	reservedRemote
	open
	halfClosedLocal
	halfClosedRemote
	closed
)

type Stream struct {
	id        		uint32
	streamDependency uint32
	priorityWeight 	byte
	lastFrame *frame.Frame
	state     StreamState
	headers   []byte
	data      []byte
}

// Build builds and returns a Request based on recieved headers and data frames
func (s *Stream) toRequest(context *hpack.Context) (*http.Request, error) {
	// Merge and Decode headers
	decoded, err := context.Decode(s.headers) // Header decompression
	if err != nil {
		return nil, err
	}

	// Build request
	req := http.NewRequest()

	// Parse Headers
	for _, hf := range decoded {
		if strings.HasPrefix(hf.Name, ":") {
			parsePseudoHeader(req, hf.Name, hf.Value)
		} else {
			req.Header[hf.Name] = hf.Value
		}
	}

	// Set body
	req.Body = s.data

	return req, nil
}

// ------- HELPERS ---------

// Parses HTTP2 Psuedo-Request-Header fields that starts with ":".
func parsePseudoHeader(req *http.Request, headerName string, value string) {
	switch headerName {
	case ":authority":
		req.Authority = value
	case ":method":
		req.Method = value
	case ":path":
		uriValues := strings.SplitN(value, "?", 2)
		req.URI = uriValues[0]
		if len(uriValues) > 1 {
			req.RawQuery = uriValues[1]
		}
	case ":scheme":
		req.Scheme = value
	}
}
