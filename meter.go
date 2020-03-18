package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

//Meter is
type (
	Meter struct {
		ID          string `json:"id" form:"id" query:"id"`
		IP          string `json:"ip" form:"ip" query:"ip"`
		Well        string `json:"well" form:"well" query:"well"`
		Level       string `json:"level" form:"level" query:"level"`
		Tag         string `json:"tag" form:"tag" query:"tag"`
		DeviceName  string `json:"devicename" form:"devicename" query:"devicename"`
		DeviceCount string `json:"devicecount" form:"devicecount" query:"devicecount"`
		DeviceKWH   string `json:"deviceKWH" form:"deviceKWH" query:"deviceKWH"`
		Ratio       string `json:"ratio" form:"ratio" query:"ratio"`
	}
)

var (
	meters = make(map[string]Meter)
)

// ReadMeters will
func readFileMeters(fname string) {
	bufms, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Error(err)
		writeFileMeters(fname)
		return
	}
	err = json.Unmarshal(bufms, &meters)
	if err != nil {
		log.Error(err)
	}

	log.Debug("meter readFileMeters: ", len(meters))
}

// WriteMeters will
func writeFileMeters(fname string) {
	bytes, err := json.Marshal(meters)
	if err != nil {
		log.Error(err)
	}
	if err := ioutil.WriteFile(fname, bytes, 0644); err != nil {
		log.Error(err)
	}
}

// GetMeterRatio 返回互感系数的整型值
func GetMeterRatio(m *Meter) float32 {
	ratio, err := strconv.Atoi(m.Ratio)
	if err != nil {
		log.Error("Atoi(m.Ratio)  ", err, m)
	}
	if ratio < 1 || ratio > 200 {
		ratio = 1.0
	}
	return float32(ratio)
}

// GetMeters [GET] 获取井表
func GetMeters(c echo.Context) error {
	var ms []Meter
	for _, m := range meters {
		ms = append(ms, m)
	}
	log.Info("GetMeters ", len(ms))
	// limit := c.Get("limit").(int)
	// offset := c.Get("offset").(int)
	// setPaginationHeader(c, limit > len(ms))
	// return c.JSON(http.StatusOK, ms[offset:offset+limit])
	return c.JSON(http.StatusOK, ms)
}

// CreateMeter [POST] 新建井表
func CreateMeter(c echo.Context) error {
	var mutex sync.Mutex
	var m = new(Meter)
	if err := c.Bind(m); err != nil {
		return err
	}
	if len(m.ID) != 12 {
		return c.NoContent(http.StatusNotAcceptable)
	}
	mutex.Lock()
	meters[m.ID] = *m
	mutex.Unlock()

	writeFileMeters(fnameMeters)
	return c.NoContent(http.StatusOK)
}

// UpdateMeter [PUT] 更新井表
func UpdateMeter(c echo.Context) error {
	var mutex sync.Mutex
	var m = new(Meter)
	if err := c.Bind(m); err != nil {
		return err
	}

	mutex.Lock()
	meters[m.ID] = *m
	mutex.Unlock()

	writeFileMeters(fnameMeters)
	return c.NoContent(http.StatusOK)
}

// DeleteMeter [DELETE] 删除井表
func DeleteMeter(c echo.Context) error {
	var mutex sync.Mutex
	id := c.Param("id")

	mutex.Lock()
	delete(meters, id)
	mutex.Unlock()

	log.Debugf("meter DeleteMeter: id=%d, meter=%v", id, meters[id])
	writeFileMeters(fnameMeters)
	return c.NoContent(http.StatusOK)
}
