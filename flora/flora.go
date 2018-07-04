package flora

import (
	"errors"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"log"
	"net"
)

const (
	ServerTypeShadowSocks = "shadowsocks"
	ServerTypeCustom      = "custom"
	ServerTypeHttp        = "http"
	ServerTypeHttps       = "https"
	ServerTypeDirect      = "direct"
	ServerTypeReject      = "direct"

	LocalServerSocksV5 = "localSocksv5"
	LocalServerHttp    = "localHttp"
)

type ProxyServer interface {
	//proxy type
	ProxyType() string
	//dial
	DialWithRawAddr(raw []byte, host string) (remote net.Conn, err error)
	//
	FailCount() int

	AddFail()
	//
	ResetFailCount()
}

type Rule struct {
	Match  string
	Action string
}

var (
	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
	errReject        = errors.New("socks reject this request")
	errSupported     = errors.New("proxy type not supported")
)

func Run(listenAddr string) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listen socket", listenAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	isClose := false
	defer func() {
		if !isClose {
			conn.Close()
		}
	}()

	var remote net.Conn

	//create remote connect
	defer func() {
		if !isClose {
			remote.Close()
		}
	}()
	go ss.PipeThenClose(conn, remote, nil)
	ss.PipeThenClose(remote, conn, nil)
	isClose = true
}
