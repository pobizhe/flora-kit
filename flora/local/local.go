package local

import (
	"errors"
	"net"
	"log"
)

const (
	socksVer5        = 5
	socksVer4        = 4
	socks5CmdConnect = 1

	socksIdVer = 0
	TypeIPv4   = 1 // type is ipv4 address
	TypeDm     = 3 // type is domain address
	TypeIPv6   = 4 // type is ipv6 address
)

var (
	errAddrType      = errors.New("socks addr type not supported")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
	errSupported     = errors.New("proxy protocol not supported")
)

type ProxyRequest struct {
	Local    net.Conn
	RawData  *[]byte
	Host     string
	HostType int
}

func NewProxyRequest(conn net.Conn) (pr *ProxyRequest, err error) {
	var (
		firstBuf []byte
		n        int
	)
	pr = &ProxyRequest{
		Local: conn,
	}
	firstBuf = make([]byte, 1)
	if n, err = conn.Read(firstBuf[:]); n > 0 && nil == err {
		switch firstBuf[socksIdVer] {
		case socksVer5:
			err = handshake(conn, firstBuf[socksIdVer])
			pr.Host, pr.HostType, err = socks5Connect(conn)
		case socksVer4:
			pr.Host, pr.HostType, err = socks4Connect(conn, firstBuf[socksIdVer])
		default:
			pr.Host, pr.HostType, pr.RawData, err =
				httpProxyConnect(conn, firstBuf[socksIdVer])
			if nil != err {
				//other protocol
				log.Print("the protocol is not supported")
				return nil, errSupported
			}
		}
	}
	return pr, err
}
