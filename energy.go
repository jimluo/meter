package main

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type (
	// EnergyRequest for app request
	EnergyRequest struct {
		Kind string `json:"kind" form:"kind" query:"kind" validate:"required"` //[tag|well|level]
		Name string `json:"name" form:"name" query:"name" validate:"required"`
		Days string `json:"days" form:"days" query:"days" validate:"required"`
	}
	// Energy will return to app
	Energy struct {
		Day   string  `json:"day" form:"day" query:"day"`
		Cmd   string  `json:"cmd" form:"cmd" query:"cmd"`
		Well  string  `json:"well" form:"well" query:"well"`
		Level string  `json:"level" form:"level" query:"level"`
		Tag   string  `json:"tag" form:"tag" query:"tag"`
		PA    float32 `json:"pa" form:"pa" query:"pa"`
		NA    float32 `json:"na" form:"na" query:"na"`
		PR    float32 `json:"pr" form:"pr" query:"pr"`
		AV    float32 `json:"av" form:"av" query:"av"`
		BV    float32 `json:"bv" form:"bv" query:"bv"`
		CV    float32 `json:"cv" form:"cv" query:"cv"`
		AA    float32 `json:"aa" form:"aa" query:"aa"`
		BA    float32 `json:"ba" form:"ba" query:"ba"`
		CA    float32 `json:"ca" form:"ca" query:"ca"`
	}
	// PowerConsumption will return to app
	PowerConsumption struct {
		DayStart   string    `json:"daystart" form:"daystart" query:"daystart"`
		DayStop    string    `json:"daystop" form:"daystop" query:"daystop"`
		TagKey     []string  `json:"tagkey" form:"tagkey" query:"tagkey"`       // 3 kinds of tag
		TagValue   []float32 `json:"tagvalue" form:"tagvalue" query:"tagvalue"` // 3 kinds of tag
		LevelKey   []string  `json:"levelkey" form:"levelKey" query:"levelkey"`
		LevelValue []float32 `json:"levelvalue" form:"levelvalue" query:"levelvalue"`
		WellKey    []string  `json:"wellkey" form:"wellkey" query:"wellkey"`       // 3 kinds of tag
		WellValue  []float32 `json:"wellvalue" form:"wellvalue" query:"wellvalue"` // 3 kinds of tag
	}

	// ConsumptionRequest will
	ConsumptionRequest struct {
		DayStart string `json:"daystart" form:"daystart" query:"daystart"`
		DayStop  string `json:"daystop" form:"daystop" query:"daystop"`
		Wells    string `json:"wells" form:"wells" query:"wells"`
	}
)

var (
	db         *leveldb.DB
	es         = []Energy{}
	cmdTimeIdx = map[string]int{}
)

// InitDB open load db
func InitDB() {
	var err error
	if db, err = leveldb.OpenFile("./db", nil); err != nil {
		log.Error("db open error", err, db)
		panic(err)
	}
	// defer db.Close()

	s := new(leveldb.DBStats)
	db.Stats(s)
	log.Info("db LevelSizes ", s.LevelSizes, "LevelTablesCounts ", s.LevelTablesCounts)
	log.Infof("BlockCacheSize %d, OpenedTablesCount %d ", s.BlockCacheSize, s.OpenedTablesCount)
	// log.Infof("AliveSnapshots %d, AliveIterators %d ", s.AliveSnapshots, s.AliveIterators)
}

// SaveDB will
func SaveDB() {
	for {
		idtimecmddata := <-chMeter // id12 time8 cmd2 data2
		db.Put(idtimecmddata[:22], idtimecmddata[22:24], nil)
		// log.Info("SaveDB ", idtimecmddata)
	}
}

func cmdToData(e *Energy, cmd uint16, d float32, r float32) {
	i := Search(cmd)
	// rd := float32(rand.Intn(60))
	// if d == 0 {
	// 	r = 1.0
	// 	switch i {
	// 	case 0, 1:
	// 		d = (3700 + rd) * 1000.0
	// 	case 2:
	// 		d = 900 + rd
	// 	case 3, 4, 5:
	// 		d = 225 + rd
	// 	case 6, 7, 8:
	// 		d = (45 + rd/10) * 100
	// 	}
	// }
	// log.Debugf("Energy cmdToData Search %d %b", i, cmd)
	switch i {
	case 0:
		e.PA = d * r / 1000.0
	case 1:
		e.NA = d * r / 1000.0 // 28 01 01 01 = 010101.28  101
	case 2:
		e.PR = d / 1000.0 // 99 09 = 0.999  09*100+99=999
	case 3:
		e.AV = d
	case 4:
		e.BV = d
	case 5:
		e.CV = d
	case 6:
		e.AA = d * r / 100.0 // 99 04 = 4.99
	case 7:
		e.BA = d * r / 100.0
	case 8:
		e.CA = d * r / 100.0
	}
}

func parseDB(k []byte, v []byte) {
	id, time := string(k[:12]), string(k[12:20])
	cmd := binary.BigEndian.Uint16(k[20:22])
	idtime := id + time
	d := float32(binary.BigEndian.Uint16(v))

	m := Meter{Well: "00000000", Ratio: "1"}
	if mOK, ok := meters[id]; ok {
		m = mOK
	} else {
		log.Error("parseDB not found id=", id)
	}
	ratio := GetMeterRatio(&m)
	// log.Debugf("GetEnergy %s %s %d %d", id, m.Well, cmd, d)
	// log.Debugf("GetEnergy id=%s time=%s d=%f cmd=%d", id, time, d, cmd)
	if _, ok := cmdTimeIdx[idtime]; !ok {
		e := Energy{Day: time, Well: m.Well, Tag: m.Tag, Level: m.Level}
		cmdTimeIdx[idtime] = len(es)
		cmdToData(&e, cmd, d, ratio)
		es = append(es, e)
		// log.Debugf("energy cmdTimeIdx OK: %s %d", idtime, len(cmdTimeIdx))
	} else {
		idx := cmdTimeIdx[idtime]
		cmdToData(&es[idx], cmd, d, ratio)
		// log.Debugf("energy cmdTimeIdx !!: %s %d %d", idtime, idx, len(es))
	}
}

func makeEnergy(w string) {
	es = []Energy{} // global, set data at parseDB()!!!
	cmdTimeIdx = map[string]int{}

	t := time.Now()
	days := []string{t.Format("20060102")}
	if w == "" {
		days = append(days, t.Format("20060102"))
	}
	if w != "" { // chart Line
		days = []string{}
		for i := 0; i < 10; i++ {
			days = append(days, t.AddDate(0, 0, (i-10)).Format("20060102"))
		}
	}
	// log.Debug("makeEnergy: days=", days)
	for _, m := range meters {
		if w != "" && m.Well != w {
			continue
		}
		// log.Debugf("GetEnergy 0: w=%s ", w)
		for _, d := range days {
			for _, c := range cmds {
				k := append([]byte(m.ID+d), byte(c>>8), byte(c))
				v, err := db.Get(k, nil)
				if err != nil {
					v = []byte{0, 0}
				}
				parseDB(k, v)
				// log.Debugf("GetEnergy 00: w=%s d=%s k=%v v=%v", w, d, k, v)
			}
		}
	}
	// log.Debug("GetEnergy levelDB err: ", iter.Error())
	log.Debug("GetEnergy 1", w, es)
	return
}

// GetEnergy will response browser,url littlecase,
func GetEnergy(c echo.Context) error {
	makeEnergy("")
	log.Debug("GetEnergy 2", es)
	return c.JSON(http.StatusOK, es)
}

// GetStatsByWell will response browser,url littlecase,
func GetStatsByWell(c echo.Context) error {
	makeEnergy(c.Param("well"))
	log.Debug("GetStatsByWell ", es)
	return c.JSON(http.StatusOK, es)
}

func getPA(id string, day string) float32 {
	k := append([]byte(id+day), 0x43, 0xc3)
	v, err := db.Get(k, nil)

	if err == nil {
		return float32(binary.BigEndian.Uint16(v))
	}

	return 0 //1000.0 + float32(rand.Intn(1000))/10.0
}

func makePowerConsumption(start string) PowerConsumption {
	dayStart, dayStop := start, time.Now().Format("20060102")
	tag, level, well := make(map[string]float32), make(map[string]float32), make(map[string]float32)
	pc := PowerConsumption{DayStart: dayStart, DayStop: dayStop}

	for _, m := range meters {
		d := getPA(m.ID, dayStop) - getPA(m.ID, dayStart)
		if d < 0 {
			d = 10.0
		}
		well[m.ID] = d
		tag[m.Tag] += d
		level[m.Level] += d
	}

	for k, v := range tag {
		pc.TagKey = append(pc.TagKey, k)
		pc.TagValue = append(pc.TagValue, v)
	}
	for k, v := range level {
		pc.LevelKey = append(pc.LevelKey, k)
		pc.LevelValue = append(pc.LevelValue, v)
	}
	// iwell := 0
	for k, v := range well {
		pc.WellKey = append(pc.WellKey, meters[k].Well) //strconv.Itoa(iwell))
		pc.WellValue = append(pc.WellValue, v)
		// iwell++
	}

	return pc
}

// GetPowerConsumption statistic all data and genereate report xls file
func GetPowerConsumption(c echo.Context) error {
	dayStart := c.Param("start")
	if len(dayStart) != 8 {
		return c.JSON(http.StatusNoContent, nil)
	}
	pc := makePowerConsumption(dayStart)
	makeReport(dayStart)

	log.Debug("GetPowerConsumption: ", pc)
	return c.JSON(http.StatusOK, pc)
}

// makeReport make diff report for download
func makeReport(start string) {
	pc := makePowerConsumption(start)

	totalPC := float32(0)
	for _, v := range pc.TagValue {
		totalPC += v
	}
	title := map[string]string{"C1": "井组电表计量系统-功耗统计", "A3": "日期区间", "A4": "总功耗(KWH)", "A6": "类型合计", "A9": "井区合计", "A12": "单井"}
	total := map[string]string{"B3": pc.DayStart + "~" + pc.DayStop, "B4": fmt.Sprintf("%f", totalPC)}

	f := excelize.NewFile()
	for k, v := range title {
		f.SetCellValue("Sheet1", k, v)
	}

	for k, v := range total {
		f.SetCellValue("Sheet1", k, v)
	}

	iCol := rune('A')
	for i, v := range pc.TagValue {
		iCol++
		f.SetCellValue("Sheet1", string(iCol)+"6", pc.TagKey[i])
		f.SetCellValue("Sheet1", string(iCol)+"7", v)
	}

	iCol = rune('A')
	for i, v := range pc.LevelValue {
		iCol++
		f.SetCellValue("Sheet1", string(iCol)+"9", pc.LevelKey[i])
		f.SetCellValue("Sheet1", string(iCol)+"10", v)
	}

	iRow := 13
	for i, v := range pc.WellValue {
		kk := strconv.Itoa(iRow)
		iRow++
		f.SetCellValue("Sheet1", "A"+kk, pc.WellKey[i])
		f.SetCellValue("Sheet1", "B"+kk, v)
	}

	err := f.AddChart("Sheet1", "D12", `{
		"type":"pie","series":[{"categories":"Sheet1!$B$6:$E$6","values":"Sheet1!$B$7:$E$7"}],
		"title":{"name":"类型合计"},"plotarea":{"show_bubble_size":true,"show_percent":true}}`)
	if err != nil {
		log.Error(err)
	}

	err = f.AddChart("Sheet1", "I12", `{
		"type":"col","series":[{"categories":"Sheet1!$B$9:$D$9","values":"Sheet1!$B$10:$D$10"}],
		"legend":{"position":"left"},"title":{"name":"井区合计"},"plotarea":{"show_val":true}}`)
	if err != nil {
		log.Error(err)
	}

	err = f.AddChart("Sheet1", "D28", `{
		"type":"col","series":[{"categories":"Sheet1!$A$13:$B$50","values":"Sheet1!$B$13:$B$50"}],
		"legend":{"position":"left"},"title":{"name":"单井功耗"}}`)
	if err != nil {
		log.Error(err)
	}

	// Save xlsx file by the given path.
	err = f.SaveAs("./井组电量功耗统计.xlsx")
	if err != nil {
		log.Error(err)
	}
}

// func makeDays() (days []string) {
// 	t := time.Now()
// 	month := t.AddDate(0, -6, -t.Day()).Format("20060102150405")[:8]

// 	iter := db.NewIterator(util.BytesPrefix([]byte(month)), nil)
// 	k := []byte{}
// 	for iter.Next() {
// 		k = iter.Key()
// 		break
// 	}
// 	iter.Release()

// 	if len(k) > 8 {
// 		for i := 0; i < 12; i++ {
// 			days = append(days, t.AddDate(0, -6, -t.Day()).Format("20060102"))
// 		}
// 	} else {
// 		for i := 0; i < 12; i++ {
// 			days = append(days, t.AddDate(0, 0, -i).Format("20060102"))
// 		}
// 	}
// 	return
// }

// GetStats statistic all data and genereate report xls file
// func GetStats(c echo.Context) error {
// 	// ss := []Stats{}
// 	days := makeDays()

// 	for i, tag := range []string{"采油", "注水", "集输"} {
// 		for _, m := range GetMetersBy("tag", tag) {
// 			for j, d := range days {
// 				for k, c := range cmds {
// 					key := append([]byte(m.Well+d), byte(c), byte(c>>8))
// 					v, err := db.Get(key, nil)
// 					if err != nil {
// 						v = []byte{0, 0}
// 					}
// 					k = k / 3
// 					stats[i][j][k] += binary.BigEndian.Uint16(v)
// 				}
// 			}
// 		}
// 	}

// 	return c.JSON(http.StatusOK, es)
// }
