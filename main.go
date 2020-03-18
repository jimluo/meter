package main

import (
	"flag"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

// 定义全区变量 为了保证执行顺序 初始化均在main中执行
var (
	chMeter = make(chan []byte)
	done    = make(chan bool)
	ticker  *time.Ticker
)

const (
	fnameUsers  = "./dbuser"
	fnameMeters = "./dbmeter"
)

// @title 井组电表计量系统
// @version 1.0
// @contact.name 罗进
// @contact.email luojinlj@gmail.com
// TODO: https, report user select date
func main() {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Errorf("panic: ", err) // 这里的err其实就是panic传入的内容
			debug.PrintStack()
		}
	}()

	isMock := flag.Bool("mock", false, "mock server for localhost test")
	isTestCmdIP := flag.String("cmd", "", "send cmd to meter for test")
	flag.Parse()
	if *isMock {
		meters = MockServers()
	}
	if *isTestCmdIP != "" {
		log.Debugf("%s  %s %s", *isTestCmdIP, flag.Arg(0), flag.Arg(1))
		if idx, err := strconv.Atoi(flag.Arg(1)); err != nil {
			testMeterCmd(*isTestCmdIP, flag.Arg(0), idx)
		}
		return
	}

	readFileUsers(fnameUsers)
	readFileMeters(fnameMeters)

	InitDB()
	log.Debug("meter config.yml: ", config)

	ticker = time.NewTicker(time.Duration(config.Interval) * time.Second)
	time.Sleep(time.Duration(config.Delay) * time.Hour)

	go requestAllServersByTimer()
	go SaveDB()

	go startWebServer()

	<-done
	ticker.Stop()
	log.Info("Client request ticker stopped!")
}

func requestAllServersByTimer() {
	for {
		select {
		// case <-done:
		// 	return
		case <-ticker.C:
			for _, m := range meters {
				if !strings.Contains(m.IP, ":") {
					m.IP = m.IP + ":4196"
				}
				go RequestDTL645(m.IP, m.ID)
			}

			// log.Info("Client request ", t.String())
		}
	}
	log.Error("requestAllServersByTimer had quit!")
}

func startWebServer() {
	// init echo
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=\"${time_rfc3339}\", remote_ip=\"${remote_ip}\", method=\"${method}\", uri=\"${uri}\", status=\"${status}\"\n",
		Output: &logWriter,
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	// e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
	// 	Skipper:   skipper,   // 跳过验证条件 在 user.go 定义
	// 	Validator: validator, // 处理验证结果 在 user.go 定义
	// }))
	e.Use(ParsePagination) // 分页参数解析，在 pagination.go 定义

	// Echo debug setting
	if config.Debug {
		e.Debug = true
	}

	// meter Routes
	e.GET("/meter", GetMeters)
	e.POST("/meter", CreateMeter)
	e.DELETE("/meter/:id", DeleteMeter)
	e.PUT("/meter", UpdateMeter)

	// user Routes
	e.POST("/login", Login)     //创建新的会话（登入）
	e.DELETE("/logout", Logout) // 销毁当前会话（登出）

	e.GET("/user", GetUsers)
	e.POST("/user", CreateUser) //注册
	e.PUT("/user", UpdateUser)
	e.DELETE("/user/:id", DeleteUser) //注销

	// GET /energy?level
	e.GET("/energy", GetEnergy)
	e.GET("/stats/:well", GetStatsByWell)
	e.GET("/consumption/:start", GetPowerConsumption)

	// e.Static("/images", "public/images/")
	e.Static("/", "public/")

	e.GET("/report", func(c echo.Context) error {
		return c.Attachment("井组电量功耗统计.xlsx", "井组电量功耗统计.xlsx")
	})

	// for https tls test
	// e.GET("/request", func(c echo.Context) error {
	// 	req := c.Request()
	// 	format := `
	// 	  <code>
	// 		Protocol: %s<br>
	// 		Host: %s<br>
	// 		Remote Address: %s<br>
	// 		Method: %s<br>
	// 		Path: %s<br>
	// 	  </code>
	// 	`
	// 	return c.HTML(http.StatusOK, fmt.Sprintf(format, req.Proto, req.Host, req.RemoteAddr, req.Method, req.URL.Path))
	// })

	// Start echo server
	e.Logger.Fatal(e.Start(":80"))
	// e.StartTLS(":443", "cert.pem", "key.pem")
}
