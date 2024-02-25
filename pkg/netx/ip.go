package netx

import "net"

// GetOutboundIP 获得对外发送消息的 IP 地址
func GetOutboundIP() string {
	// 通过 UDP 协议创建一个连接到远程地址 8.8.8.8:80 的连接
	// 这里选择了 Google 的 DNS 服务器地址
	// 国内就使用 114.114.114.114
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
