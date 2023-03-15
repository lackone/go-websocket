package go_websocket

import (
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/netip"
)

// 获取内网IP
func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addrPort, err := netip.ParseAddrPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}
	return addrPort.Addr().String(), nil
}

// 获取外网IP
func GetExternalIP() (string, error) {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(all), nil
}

// 获取远程IP
func RemoteIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		addr := req.RemoteAddr
		addrPort, err := netip.ParseAddrPort(addr)
		if err != nil {
			return ""
		}
		ip = addrPort.Addr().String()
	}
	return ip
}

// 整数IP转字符串
func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

// 字符串IP转整数
func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}
