package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
正向有功总电能  DTL645 97版，数据长度2，命令和地址先低后高
68 03 00 00 00 00 00 00 68 01 02 43 C3 DC 16
68 03 00 00 00 00 00 00 00 00 00 68 81 06 43 C3 5b 34 33 33 33 55 16
反向有功总电能
68 03 00 00 00 00 00 00 00 00 68 01 02 53 C3 EC 16
68 03 00 00 00 00 00 00 00 00 00 68 81 06 53 C3 5b 34 33 33 33 65 16

读表地址命令  68 AA AA AA AA AA AA 68 01 02 65 F3 27 16 获取电表地址

A相电流
68 03 00 00 00 00 00 68 01 02 54 E9 13 16
68 03 00 00 00 00 00 68 81 04 54 E9 CC 37 98 16
瞬时有功功率
68 03 00 00 00 00 68 01 02 63 E9 22 16
68 03 00 00 00 00 00 68 81 05 63 E9 94 B9 33 25 16
*/
const (
	commandCount     = 9
	idxStart         = 0
	idxAddr          = 1
	idxDataStart     = 7
	idxControl       = 8
	idxDataLen       = 9
	idxCmd           = 10
	idxData          = 12
	idxSum           = 12
	idxStop          = 13
	numRecvFrameByte = 24
)

var (
	// 	positiveActive, negativeActive, powerFactor,aVoltage, bVoltage, cVoltage, aCurrent, bCurrent, cCurrent
	cmds = []uint16{0x43c3, 0x53c3, 0x83e9, 0x44e9, 0x45e9, 0x46e9, 0x54e9, 0x55e9, 0x56e9}

	cacheFrame = []byte{
		0x68, //0,start Frame
		0x27, //1,Address 0 18030927
		0x09, //2,Address 1
		0x03, //3,Address 2
		0x18, //4,Address 3
		0x00, //5,Address 4
		0x00, //6,Address 5
		0x68, //7,start data,
		0x01, //8,controlRequest,01sent 81receive
		0x02, //9,data length,
		0x00, //10,command 0
		0x00, //11,command 1
		0x00, //12,Parity code
		0x16, //13,endFrame
	}

	connNet        = net.Dial
	readNet        = net.Conn.Read
	writeNet       = net.Conn.Write
	errMockNetExit = errors.New("test failure. mock readNet.for exit test")
	errWriteNet    = errors.New("Write Net timeout or failure")
	errReadNet     = errors.New("Read Net timeout or failure")
	errParse645    = errors.New("Parse 645 frame failure")
)

// Search binsearch index of x in cmds bytes
func Search(x uint16) int {
	for i, c := range cmds {
		if x == c {
			return i
		}
	}
	return 0
}

func sum645(f []byte) byte {
	sum := byte(0)
	for _, b := range f {
		sum += b
	}
	return sum
}

func setFrameField(f []byte, idx int, s string) {
	for i := 0; i < len(s); i = i + 2 {
		b, _ := strconv.ParseUint(s[i:i+2], 16, 8)
		f[idx+i/2] = byte(b)
	}
}

func makeRequestCommands(id string) (c [commandCount][]byte) {
	for i, cmd := range cmds {
		setFrameField(cacheFrame, idxAddr, id[:10])
		cacheFrame[idxCmd+1] = byte(cmd)
		cacheFrame[idxCmd] = byte(cmd >> 8)
		cacheFrame[idxSum] = sum645(cacheFrame[idxStart:idxSum])
		c[i] = append([]byte{0xFE, 0xFE}, cacheFrame...)
		// log.Debug("setFrameField ", c[i])
	}

	return
}

func hexStrData(hex byte) uint16 {
	hex -= 0x33
	shex := fmt.Sprintf("%02x", hex)
	h, err := strconv.Atoi(shex)
	if err != nil {
		return 0
	}
	return uint16(h)
}

// var id12time8cmd2data2 bytes.Buffer
// id12time8cmd2data2.WriteString(fmt.Sprintf("%02x%02x%02x%02x%02x%02x", f[idxAddr],f[idxAddr+1],f[idxAddr+2],f[idxAddr+3],f[idxAddr+4],f[idxAddr+5]))
// id12time8cmd2data2.WriteString(time.Now().Format("20060102")) //len 8 bytes
// id12time8cmd2data2.Write(f[idxCmd : idxCmd+2])
// dataLen := int(f[idxDataLen]) - lencmd
// d := uint32(0)
// for i := idxData; i < dataLen; i++ {
// 	d += uint32(i*100) * uint32(f[i]-0x33)
// }
// id12time8cmd2data2.Write([]byte{byte(d >> 8), byte(d)})
// fmt.Printf("parse645: id12time8cmd2data2=%v dataLen=%d d=%d", hex.EncodeToString(id12time8cmd2data2.Bytes()), dataLen, d)
func parse645(frame []byte) []byte {
	// 68 frame start, 16 frame end;
	// 68 data start, 81 response control
	startf := -1
	for i := 0; i < len(frame)/2; i++ {
		if frame[i] == 0x68 && frame[i+idxControl] == 0x81 { // && frame[i+idxStop+lenCmd] == 0x16 {
			startf = i
			break
		}
	}
	if startf == -1 || len(frame)-startf < 16 {
		log.Errorf("parse645 68 and len error. frame= %v", frame)
		return nil
	}

	f := frame[startf:]
	id := f[idxAddr : idxAddr+6]
	cmd := f[idxCmd : idxCmd+2]
	dataLen := int(f[idxDataLen])
	t := []byte(time.Now().Format("20060102")) //len 8 bytes

	idxD := idxData + dataLen - 4
	d1 := hexStrData(f[idxD+1]) * 100
	d2 := hexStrData(f[idxD])
	d := d1 + d2
	if dataLen == 6 {
		d = d*10000 + hexStrData(f[idxD-1])*100 + hexStrData(f[idxD-2])
	}
	dd := make([]byte, 2)
	binary.BigEndian.PutUint16(dd, d)

	ids := fmt.Sprintf("%02x%02x%02x%02x%02x%02x", id[0], id[1], id[2], id[3], id[4], id[5])
	// idtimecmddata = // 9904001800 2019111815 PA 0201 id6 t10 cmd2 data2
	// log.Debugf("cmd: %s %s", cmdstr, cmd)
	log.Debugf("parse645: id=%v cmd=%v dataLen=%d d=%d dd=%v", hex.EncodeToString(id), hex.EncodeToString(cmd), dataLen, d, hex.EncodeToString(dd))
	return bytes.Join([][]byte{[]byte(ids), t, cmd, dd}, []byte(""))
}

// net.Conn.Read
func waitResponseDTL645(conn net.Conn) error {
	frame := make([]byte, 24)
	n, err := readNet(conn, frame)
	if err != nil {
		return fmt.Errorf("Client recv error %d %s %v %w", n, err, frame, errReadNet)
	}
	idtimecmddata := parse645(frame)
	if idtimecmddata == nil || len(idtimecmddata) < 24 {
		return errParse645
	}
	// log.Debug("Rx: ", idtimecmddata)

	chMeter <- idtimecmddata

	return nil
}

// RequestDTL645 用id生成多个请求命令，针对ip发出请求
func RequestDTL645(ip string, id string) {
	errNetCmd := [][]byte{}

	// log.Debug("connNet: ", ip, id)
	conn, err := connNet("tcp", ip)
	// conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	// conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Error("connNet: ", conn, err)
		return
	}

	for _, cmd := range makeRequestCommands(id) {
		// log.Debug("Tx: ", cmd)
		if n, err := writeNet(conn, cmd); err != nil {
			if err == errMockNetExit {
				return
			}
			log.Error("writeNet: ", n, err)
			errNetCmd = append(errNetCmd, cmd)
			continue
		}
		if err := waitResponseDTL645(conn); err != nil {
			errNetCmd = append(errNetCmd, cmd)
			log.Error("ReadNet: ", err)
		}
	}

	requestTryagain(ip, conn, errNetCmd)
}

// requestTryagain RequestDTL645调用后失败的conn或write read，重来一次
func requestTryagain(ip string, conn net.Conn, errNetCmd [][]byte) {
	if conn == nil {
		conn, err := connNet("tcp", ip)
		if err != nil {
			log.Error("connNet: ", conn, err)
			return
		}
	}

	for _, cmd := range errNetCmd {
		log.Debug("Tx Retry: ", cmd)
		if n, err := writeNet(conn, cmd); err != nil {
			log.Error("writeNet: ", n, err)
			continue
		}
		if err := waitResponseDTL645(conn); err != nil {
			log.Error("ReadNet: ", err)
		}
	}

	if conn != nil {
		conn.Close()
		conn = nil
	}
}

func testMeterCmd(ip string, id string, idxCmd int) {
	conn, err := connNet("tcp", ip)
	if err != nil {
		log.Error("connNet: ", conn, err)
		return
	}
	defer conn.Close()

	go waitResponseDTL645(conn)

	cmds := makeRequestCommands(id)
	for i := 0; i < 10; i++ {
		if n, err := writeNet(conn, cmds[idxCmd]); err != nil {
			log.Error("writeNet: ", n, err)
		}
		time.Sleep(1 * time.Second)
	}
}
