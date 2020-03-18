package main

import (
	"testing"
)

func TestParseDB(t *testing.T) {
	es = []Energy{}
	cmdTimeIdx = map[string]int{}
	k, v := []byte{50, 54, 48, 57, 48, 51, 49, 56, 48, 48, 48, 48, 50, 48, 49, 57, 49, 50, 49, 54, 49, 57, 67, 195}, []byte{0, 0}
	for _, c := range cmds {
		k[23] = byte(c)
		k[22] = byte(c >> 8)
		v[1]++

		parseDB(k, v)
	}

	t.Error(cmdTimeIdx, es)
	// k, v = []byte{50, 54, 48, 57, 48, 51, 49, 56, 48, 48, 48, 48, 50, 48, 49, 57, 49, 50, 49, 54, 49, 57, 67, 195}, []byte{0, 10}
	// parseDB(k, v, cmdTimeIdx, es)
	// t.Error(cmdTimeIdx, es)
	// id, time := string(k[:12]), string(k[12:22])
	// cmd := binary.BigEndian.Uint16(k[22:24])
	// idtime := id + time
	// d := binary.BigEndian.Uint16(v)

	// m := meters[id]
	// ratio := GetMeterRatio(&m)
	// // log.Debugf("GetEnergy %s %s %d %d", id, m.Well, cmd, d)
	// t.Errorf("GetEnergy id=%s time=%s d=%d cmd=%d", id, time, d, cmd)
	// if _, ok := cmdTimeIdx[idtime]; !ok {
	// 	e := new(Energy)
	// 	cmdTimeIdx[idtime] = len(es)
	// 	cmdToData(e, cmd, d, ratio)
	// 	e.Day = time
	// 	e.Well = m.Well
	// 	e.Tag = m.Tag
	// 	e.Level = m.Level
	// 	es = append(es, *e)
	// 	t.Error("energy: ", idtime, len(cmdTimeIdx), ok, es)
	// } else {
	// 	idx := cmdTimeIdx[idtime]
	// 	cmdToData(&es[idx], cmd, d, ratio)
	// 	t.Error("energy: ", idtime, es)
	// }

	// k, v = []byte{50, 54, 48, 57, 48, 51, 49, 56, 48, 48, 48, 48, 50, 48, 49, 57, 49, 50, 49, 54, 49, 57, 83, 195}, []byte{0, 11}
	// id, time = string(k[:12]), string(k[12:22])
	// cmd = binary.BigEndian.Uint16(k[22:24])
	// idtime = id + time
	// d = binary.BigEndian.Uint16(v)

	// m = meters[id]
	// ratio = GetMeterRatio(&m)
	// // log.Debugf("GetEnergy %s %s %d %d", id, m.Well, cmd, d)
	// t.Errorf("GetEnergy id=%s time=%s d=%d cmd=%d", id, time, d, cmd)
	// if _, ok := cmdTimeIdx[idtime]; !ok {
	// 	e := new(Energy)
	// 	cmdTimeIdx[idtime] = len(es)
	// 	cmdToData(e, cmd, d, ratio)
	// 	e.Day = time
	// 	e.Well = m.Well
	// 	e.Tag = m.Tag
	// 	e.Level = m.Level
	// 	es = append(es, *e)
	// 	t.Error("energy: ", idtime, len(cmdTimeIdx), ok, es)
	// } else {
	// 	idx := cmdTimeIdx[idtime]
	// 	cmdToData(&es[idx], cmd, d, ratio)
	// 	t.Error("energy: ", idtime, es)
	// }
}
