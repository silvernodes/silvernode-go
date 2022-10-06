package netutil

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

func GetLocalIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	ips = append(ips, "127.0.0.1")
	return ips, nil
}

func IsLocalIPv4(ip string) bool {
	ips, err := GetLocalIPv4s()
	if err != nil {
		return false
	}
	for _, item := range ips {
		if item == ip {
			return true
		}
	}
	return false
}

func ParseUrlInfo(url string) (string, string, uint64, error) {
	theUrl := strings.Split(url, "://") // trim the ws header
	if len(theUrl) < 2 {
		return "", "", 0, errutil.New("非法的url地址:" + url)
	}
	infos := strings.Split(theUrl[1], "/") // parse the sub path
	ipAndPort := strings.Split(infos[0], ":")
	if len(ipAndPort) < 2 {
		return "", "", 0, errutil.New("非法的ipAddress:" + url)
	}
	port, err := strconv.ParseUint(ipAndPort[1], 10, 64)
	if err != nil {
		return "", "", 0, errutil.New("非法的端口:" + url)
	}
	return theUrl[0], ipAndPort[0], port, nil
}

func GetAvailablePort() (uint64, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return uint64(listener.Addr().(*net.TCPAddr).Port), nil
}
