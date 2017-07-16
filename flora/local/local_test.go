package local

import (
	"testing"
	"net"
	"log"
	"time"
	"os"
	"net/http"
	"net/url"
	"io"
	"sync/atomic"
)

var (
	socks4conn           = []byte{0x04, 0x01, 0x00, 0x50, 0x3a, 0xff, 0xad, 0x2e, 0x00}
	socks4connResp       = []byte{0x00, 0x5a, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00}
	socks5init           = []byte{0x05, 0x01, 0x00}
	socks5initResp       = []byte{0x05, 0x00}
	socks5conn           = []byte{0x05, 0x01, 0x00, 0x01, 0x3a, 0xff, 0xad, 0x2e, 0x00, 0x50}
	socks5connResp       = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43}
	sigs                 = make(chan os.Signal, 1)
	port                 = "3000"
	start          int32 = 0
)

func testListen() {
	if start > 0 {
		return
	} else {
		atomic.AddInt32(&start, 1)
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("accept:", err)
			continue
		}
		go handleConnection(conn)
	}
	<-sigs
}

func handleConnection(conn net.Conn) {
	isClose := false
	defer func() {
		if !isClose && nil != conn {
			conn.Close()
		}
	}()
	if req, err := NewProxyRequest(conn); nil == err {
		remote, err := net.Dial("tcp", req.Host)
		if nil != err {
			//
		}
		defer func() {
			if !isClose && nil != conn {
				conn.Close()
			}
		}()
		if nil != req.RawData && len(*req.RawData) > 0 {
			remote.Write(*req.RawData)
		}
		go io.Copy(remote, req.Local)
		io.Copy(req.Local, remote)
		isClose = true
	}
}

func TestNewProxyRequest(t *testing.T) {
	t.Log("start listen")
	go testListen()
	testProxy(t)
}

func testProxy(t testing.TB) {
	time.Sleep(time.Duration(4000))
	testSocks4(t)
	testSocks5(t)
	testHttpProxy(t)
}

func BenchmarkNewProxyRequest(t *testing.B) {
	go testListen()
	//time.Sleep(time.Duration(5000))
	for i := 0; i < t.N; i++ {
		testProxy(t)
	}
}

func testHttpProxy(t testing.TB) {
	proxyUrl, _ := url.Parse("http://127.0.0.1:" + port)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	req, err := http.NewRequest("GET", "http://www.cnbeta.com", nil)
	if err != nil {
		t.Fatal("make request has error!")
	}
	resp, err := client.Do(req)
	if nil != err {
		t.Fatal("http proxy has error")
	}
	if resp.StatusCode == http.StatusOK {
		t.Log("test http proxy pass")
	} else {
		t.Fatal("http proxy NG")
	}
}

func testSocks5(t testing.TB) {
	conn, _ := net.Dial("tcp", ":"+port)
	defer func() {
		if nil != conn {
			conn.Close()
		}
	}()
	conn.Write(socks5init)
	buf := make([]byte, len(socks5initResp))
	if !testReadBytes(conn, socks5initResp[:], buf[:]) {
		t.Log("test socks5init Right ", socks5initResp)
		t.Fatal("test socks5init NG ", buf)
	}
	buf = make([]byte, len(socks5connResp))
	conn.Write(socks5conn)
	if !testReadBytes(conn, socks5connResp[:], buf[:]) {
		t.Log("test socks5connResp Right ", socks5initResp)
		t.Fatal("test socks5connResp NG ", buf)
	}
	t.Log("test socks5 pass ")
}

func testSocks4(t testing.TB) {
	conn, err := net.Dial("tcp", ":"+port)
	if nil != err {
		t.Fatal("conn is error")
	}
	defer func() {
		if nil != conn {
			conn.Close()
		}
	}()
	conn.Write(socks4conn)
	buf := make([]byte, len(socks4conn))
	if !testReadBytes(conn, socks4connResp[:], buf[:]) {
		t.Log("test socks4 Right ", socks4connResp)
		t.Fatal("test socks4 NG ", buf)
	}
	t.Log("test socks4 pass ")
}

func testReadBytes(conn net.Conn, defResp []byte, readBuf []byte) bool {
	if n, err := conn.Read(readBuf[:]); n > 0 || nil == err {
		for i, b := range defResp {
			if readBuf[i] != b {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
