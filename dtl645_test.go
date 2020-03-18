package main

import (
	"encoding/binary"
	"encoding/hex"
	"net"
	"testing"
)

func mockNetFunc() {
	connNet = func(t, ip string) (c net.Conn, e error) {
		return c, nil
	}
	writeNet = func(conn net.Conn, d []byte) (int, error) {
		return 0, errMockNetExit
	}

	entryReadNetCount := 0
	readNet = func(conn net.Conn, d []byte) (n int, err error) {
		if entryReadNetCount++; entryReadNetCount > 0 {
			// log.Infof("entryReadNetCount %d", entryReadNetCount)
			n, err = entryReadNetCount, errMockNetExit
		} else {
			copy(d, "FEFEFEFE6827090318000068810443C334352416")
		}
		return
	}

}

func TestSum645(t *testing.T) {
	b := []byte{0x68, 0x02, 0x20, 0x00, 0x17, 0x10, 0x01, 0x68, 0x11, 0x04, 0x33, 0x33, 0x34, 0x33}
	got := sum645(b)
	if 252 != got {
		t.Error("sum645 252 != ", got)
	}
}

func TestMakeRequestCommands(t *testing.T) {
	want := map[string][][]byte{
		"270903180000": {
			[]byte{0xFE, 0xFE, 0x68, 0x27, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x43, 0xC3, 0x24, 0x16},
			[]byte{0xFE, 0xFE, 0x68, 0x27, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x53, 0xc3, 0x34, 0x16},
			[]byte{0xFE, 0xFE, 0x68, 0x27, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x83, 0xE9, 0x8a, 0x16},
		},
		"260903180000": {
			[]byte{0xFE, 0xFE, 0x68, 0x26, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x43, 0xC3, 0x23, 0x16},
			[]byte{0xFE, 0xFE, 0x68, 0x26, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x53, 0xc3, 0x33, 0x16},
			[]byte{0xFE, 0xFE, 0x68, 0x26, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x01, 0x02, 0x83, 0xE9, 0x89, 0x16},
		},
	}
	for deviceID, v := range want {
		got := makeRequestCommands(deviceID)

		for i := 0; i < 3; i++ {
			if hex.EncodeToString(got[i]) != hex.EncodeToString(v[i]) {
				// if want != string(c) {
				t.Errorf("makeRequestCommands %d %v %v %x", i, got[i], v[i], cmds[i])
				// }
			}
		}
	}
}

func TestParse645(t *testing.T) {
	b := []byte{0xFE, 0xFE, 0x68, 0x26, 0x09, 0x03, 0x18, 0x00, 0x00, 0x68, 0x81, 0x04, 0x43, 0xC3, 33, 34, 0x23, 0x16}
	idtimecmddata := parse645(b)
	k, v := idtimecmddata[:24], idtimecmddata[24:26]
	id := string(k[:12])
	// id := fmt.Sprintf("%02d%02d%02d%02d%02d00", k[0], k[1], k[2], k[3], k[4])
	d := binary.BigEndian.Uint16(v)
	cmd := binary.BigEndian.Uint16(k[22:24])

	// t.Error(idtimecmddata)
	if id != "260903180000" || cmd != 0x43c3 || d != 10 {
		t.Errorf("parse645: %v id=%s cmd=%d d=%d", idtimecmddata, id, cmd, d)
	}
}

// func TestRequestDTL645(t *testing.T) {
// 	mockNetFunc()
// 	testIP := "123.456.789.000"
// 	RequestDTL645(testIP, "123456")
// }
