/*
A very simple TCP server written in Go.

This is a toy project that I used to learn the fundamentals of writing
Go code and doing some really basic network stuff.

Maybe it will be fun for you to read. It's not meant to be
particularly idiomatic, or well-written for that matter.
*/
package main

import (
	"net"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	host      = "127.0.0.1"
	portFirst = 9900
	portCount = 50
)

// MockServers make mock server for test
func MockServers() map[string]Meter {
	var wg sync.WaitGroup

	ms := make(map[string]Meter, portCount)
	for i := 0; i < portCount; i++ {
		port := strconv.Itoa(i + portFirst)
		ip := host + ":" + port
		id := "00" + port + port + "00"
		log.Debugf("MockServers %s %s", id, ip)
		ms[id] = Meter{ID: id, IP: ip}
		wg.Add(1)
		go connect(ip, &wg)
	}
	wg.Wait()

	log.Infof("Starting server %s:%d~%s:%d", host, portFirst, host, portFirst+portCount)

	return ms
}

func connect(ip string, wg *sync.WaitGroup) {
	listener, err := net.Listen("tcp", ip)
	if err != nil {
		log.Errorf("server %s listen error %s", ip, err)
	}
	wg.Done() //wg.Add(-1)

	defer listener.Close()

	// log.Info("server listen:", ip)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Infof("server %s connection error %s", ip, err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	data := make([]byte, 36)
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Infof("Server %s recv error %s", conn.LocalAddr(), err)
			break
		}
		// log.Infof("Server %s recv from %s data %s", conn.LocalAddr(), conn.RemoteAddr(), string(data))
		//req 68 78 56 34 12 00 00 68 11 04 33 33 34 33 C6 16
		// id := parse645(data)
		// if id == nil {
		// 	continue
		// }

		frame := []byte{
			0x68,
			11, 22, 33, 44, 55, 66,
			0x68,
			0x81,
			0x04,
			0x83, //10,command 0
			0xe9, //11,command 1
			0x33,
			0x34,
			0x00, //14,Parity code
			0x16, //15,endFrame
		}
		frame[14] = sum645(frame[:14])
		// log.Infof("Server %s sent %d %s ", frame, startf, f)
		if n > 0 {
			conn.Write(frame)
		}
		log.Infof("Server %s sent %d %x ", conn.LocalAddr(), n, frame)
	}

	log.Info("Server disconnected by Client ", conn.LocalAddr())
}
