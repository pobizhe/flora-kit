package local

import (
	"net"
	"io"
	"net/http"
	"bufio"
	"bytes"
	"net/http/httputil"
)

// local socks server  connect
func httpProxyConnect(conn net.Conn, first byte) (addr string, hostType int, raw *[]byte, err error) {
	var (
		HTTP_200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
		host     string
		port     string
		rawData  []byte
	)

	buf := make([]byte, 4096)
	buf[0] = first
	io.ReadAtLeast(conn, buf[1:], 1)
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if nil != err {
		return
	}
	host, port, err = net.SplitHostPort(req.Host)
	if nil != err {
		host = req.Host
		port = req.URL.Port()
	}
	scheme := req.URL.Scheme
	if "" == port {
		if scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}
	addr = net.JoinHostPort(host, port)
	method := req.Method
	hostType = getRequestType(addr)
	switch method {
	case http.MethodConnect:
		_, err = conn.Write(HTTP_200)
	default:
		removeProxyHeaders(req)
		rawData, err = httputil.DumpRequest(req, true)
		raw = &rawData
	}
	return
}

func getRequestType(addr string) int {
	host, _, _ := net.SplitHostPort(addr)
	ip := net.ParseIP(host)

	if nil != ip {
		if len(ip) == net.IPv4len {
			return TypeIPv4
		} else {
			return TypeIPv6
		}
	}
	return TypeDm
}

func removeProxyHeaders(req *http.Request) {
	req.RequestURI = ""
	//req.Header.Del("Accept-Encoding")
	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	//req.Header.Del("Referer")
	// Connection, Authenticate and Authorization are single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.
	//req.Header.Del("Connection")
}
