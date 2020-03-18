package main

import (
	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	// Config for config sys
	Config struct {
		Debug    bool   `default:"true"`
		Delay    uint32 `default:"0"`
		Interval uint32 `default:"20"`
	}
)

var (
	config    Config
	logWriter = lumberjack.Logger{
		Filename:   "./log/meter.log",
		MaxSize:    20, // megabytes
		MaxBackups: 30,
		MaxAge:     35,   //days
		Compress:   true, // disabled by default
	}
)

func init() {
	configor.Load(&config, "./config.yml")
	// if config.Debug {
	// 	logrus.SetFormatter(&logrus.TextFormatter{
	// 		FullTimestamp:   true,
	// 		TimestampFormat: "06-01-02 15:04:05.00",
	// 	})
	// 	logrus.SetLevel(logrus.DebugLevel)
	// } else {
	// 	logrus.SetLevel(logrus.InfoLevel)
	// }
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "06-01-02 15:04:05.00",
	})
	logrus.SetOutput(&logWriter)
}
