package request

import (
	"bytes"
	"errors"
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/headers"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       reqState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type reqState int

const (
	reqStateInitialized reqState = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

func (r reqState) String() string {
	return [...]string{"initialized", "done"}[r]
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		state:   reqStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state %d, read n bytes %d", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead
		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLine(line []byte) (*RequestLine, int, error) {
	idx := bytes.Index(line, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requetLineText := string(line[:idx])
	bytesConsumed := len(requetLineText) + len(crlf)
	requestLine, err := requestLineFromString(requetLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, bytesConsumed, nil
}

func requestLineFromString(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", parts)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]
	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", parts[2])
	}
	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func (r *Request) parseHeadersLine(data []byte) (int, bool, error) {
	n, done, err := r.Headers.Parse(data)
	if err != nil {
		return 0, false, err
	}
	return n, done, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil

}

func (r *Request) parseSingle(data []byte) (int, error) {

	switch r.state {
	case reqStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	case requestStateParsingHeaders:
		n, done, err := r.parseHeadersLine(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil

	case requestStateParsingBody:
		v := r.Headers.Get("Content-Length")
		if v == "" {
			r.state = requestStateDone
			return 0, nil
		}
		r.Body = append(r.Body, data...)
		n, err := strconv.Atoi(r.Headers.Get("Content-Length"))
		if err != nil {
			return 0, err
		}
		if len(r.Body) > n {
			return 0, fmt.Errorf("len of body greater than reported, reported %d expected %d, body %s", n, len(r.Body), string(r.Body))
		}
		if len(r.Body) == n {
			r.state = requestStateDone
			return 0, nil
		}
		return len(data), nil

	default:
		return 0, fmt.Errorf("unknown state")
	}
}
