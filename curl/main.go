package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrMethodNotSupported = errors.New("method not supported")
	allowedMethods        = []string{"GET", "DELETE", "POST", "PUT"}
	httpPort              = "80"
	httpsPort             = "443"
)

type headers struct {
	m map[string]string
}

func (h *headers) String() string {
	var str string
	for k, v := range h.m {
		str += fmt.Sprintf("%s:%s\r\n", k, v)
	}

	return str
}

func (h *headers) Set(value string) error {
	if h.m == nil {
		h.m = make(map[string]string)
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) == 2 {
		h.m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return nil
}

func main() {
	var (
		verbose   bool
		method    string
		data      string
		headersIn headers
	)

	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.StringVar(&method, "X", "GET", "HTTP method")
	flag.StringVar(&data, "d", "", "Data for POST and PUT requests")
	flag.Var(&headersIn, "H", "Headers to include with the request")

	flag.Parse()

	method = strings.ToUpper(method)

	if err := validateMethod(method); err != nil {
		panic(err)
	}

	args := flag.Args()
	if len(args) == 0 {
		panic("gimme url to curl please -_-")
	}

	w := newWriter(verbose)

	url, err := url.Parse(args[0])
	if err != nil {
		panic(err)
	}

	// w.writeIn(fmt.Sprintf("connecting to %s\n", url.Hostname()))
	w.writeIn(fmt.Sprintf("%s %s %s", method, url.RequestURI(), "HTTP/1.1"))
	w.writeIn(fmt.Sprintf("Host: %s", url.Hostname()))
	w.writeIn("Accept: */*")

	port := url.Port()
	if port == "" {
		if url.Scheme == "https" {
			port = httpsPort
		} else {
			port = httpPort
		}
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(url.Hostname(), port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write(makePayload(method, url.RequestURI(), url.Hostname(), data, headersIn))
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}
	w.writeIn("")
	w.writeOut(strings.TrimSpace(statusLine))

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		line = strings.TrimSpace(line)
		if verbose {
			w.writeOut(line)
		}
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}

	}

	contentLength := 0
	if val, ok := headers["Content-Length"]; ok {
		fmt.Sscanf(val, "%d", &contentLength)
	}

	body := make([]byte, contentLength)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		panic(err)
	}
	w.write(string(body))
}

type writer struct {
	verbose bool
}

func newWriter(verbose bool) writer {
	return writer{
		verbose: verbose,
	}
}

func (w writer) writeIn(str string) {
	if w.verbose {
		println("> " + str)
	}
}

func (w writer) writeOut(str string) {
	if w.verbose {
		println("< " + str)
	} else {
		println(str)
	}
}

func (w writer) write(str string) {
	println(str)
}

func validateMethod(str string) error {
	for _, method := range allowedMethods {
		if strings.ToUpper(str) == method {
			return nil
		}
	}

	return ErrMethodNotSupported
}

func makePayload(method, uri, host, body string, headers headers) []byte {
	fmtStr := "%s %s HTTP/1.1\r\nHost: %s\r\n"
	fmtStr += headers.String()
	if method == http.MethodPost {
		fmtStr += "Content-Length: %d\r\n" +
			"Connection: close\r\n" +
			"\r\n" +
			"%s"
	}
	fmtStr += "\r\n"

	if method == http.MethodPost {
		return []byte(fmt.Sprintf(fmtStr, method, uri, host, len(body), body))
	}

	return []byte(fmt.Sprintf(fmtStr, method, uri, host))
}

// func parseHeaders(str *string) map[string]string {
// 	if str == nil {
// 		return nil
// 	}
//
//   strings.Spl
//
// 	headerMap := make(map[string]string)
// }
