package utils

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gofiber/contrib/websocket"
)

func HandleWebSocket(c *websocket.Conn, id string, cli *client.Client, unusedContainer map[string]bool, mutex *sync.Mutex) {
	containerName := "chrome-instance-" + id
	targetAddr := containerName + ":5900"

	mutex.Lock()
	delete(unusedContainer, containerName)
	mutex.Unlock()

	vnc, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("%s: failed to bind to the VNC Server: %s", time.Now().Format(time.Stamp), err)
		// If we can't connect to VNC, clean up the container immediately
		go cleanupContainer(cli, containerName)
		return
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

	// Wait for either direction to close
	<-done

	// Clean up the container when websocket disconnects
	log.Printf("WebSocket disconnected for container %s, cleaning up...", containerName)
	go cleanupContainer(cli, containerName)
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
				return
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
		if err != nil {
			log.Printf("%s: reading from WS failed: %s", time.Now().Format(time.Stamp), err)
			return
		}

		if _, err := conn.Write(buffer); err != nil {
			log.Printf("%s: writing to TCP failed: %s", time.Now().Format(time.Stamp), err)
			return
		}
	}
}

func cleanupContainer(cli *client.Client, containerName string) {
	if cli == nil {
		log.Printf("Docker client is nil, cannot cleanup container %s", containerName)
		return
	}

	ctx := context.Background()

	// Stop the container
	timeout := 10 // 10 seconds timeout
	err := cli.ContainerStop(ctx, containerName, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		log.Printf("Failed to stop container %s: %s", containerName, err)
	} else {
		log.Printf("Container %s stopped successfully", containerName)
	}

	// Remove the container
	err = cli.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force: true, // Force removal even if running
	})
	if err != nil {
		log.Printf("Failed to remove container %s: %s", containerName, err)
	} else {
		log.Printf("Container %s removed successfully", containerName)
	}
}
