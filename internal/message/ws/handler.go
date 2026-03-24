package ws

import (
	"fmt"
	"net"
	"net/http"
	"ticktok-service/pkg/config"
)

// WsHandler handles the HTTP request and upgrades it to WebSocket
func WsHandler(w http.ResponseWriter, r *http.Request) {
	nodeAddr := getNodeAddr()
	ServeWebSocket(w, r, nodeAddr)
}

func getNodeAddr() string {
	ip := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}
	}
	return fmt.Sprintf("%s:%s", ip, config.Config.MessageService.Port)
}
