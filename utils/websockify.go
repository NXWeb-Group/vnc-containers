package utils

import (
	"log"
	"net"
	"time"

	"github.com/gofiber/contrib/websocket"
)

func HandleWebSocket(c *websocket.Conn, id string) {
	var targetAddr = "chrome-instance-" + id + ":5900"
	vnc, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("%s: failed to bind to the VNC Server: %s", time.Now().Format(time.Stamp), err)
	}

	done := make(chan bool, 2)

	go func() {
		//tcp to ws
		forwardTcp(c, vnc)
		done <- true
	}()

	go func() {
		//ws to tcp
		forwardWeb(c, vnc)
		done <- true
	}()

	<-done
}

func forwardTcp(wsConn *websocket.Conn, conn net.Conn) {
	var tcpBuffer [1024]byte

	defer func() {
		if conn != nil {
			conn.Close()

		}
		if wsConn != nil {
			wsConn.Close()
		}
	}()
	for {
		if (conn == nil) || (wsConn == nil) {
			return
		}
		n, err := conn.Read(tcpBuffer[0:])
		if err != nil {
			log.Printf("%s: reading from TCP failed: %s", time.Now().Format(time.Stamp), err)
			return
		} else {
			if err := wsConn.WriteMessage(websocket.BinaryMessage, tcpBuffer[0:n]); err != nil {
				log.Printf("%s: writing to WS failed: %s", time.Now().Format(time.Stamp), err)
			}
		}
	}
}

func forwardWeb(wsConn *websocket.Conn, conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%s: reading from WS failed: %s", time.Now().Format(time.Stamp), err)
		}
		if conn != nil {
			conn.Close()
		}
		if wsConn != nil {
			wsConn.Close()
		}
	}()
	for {
		if (conn == nil) || (wsConn == nil) {
			return
		}

		_, buffer, err := wsConn.ReadMessage()
		if err == nil {
			if _, err := conn.Write(buffer); err != nil {
				log.Printf("%s: writing to TCP failed: %s", time.Now().Format(time.Stamp), err)
			}
		}
	}
}
