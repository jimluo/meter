package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type (
	// Token 以上第四步返回给客户端的token对象
	Token struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
		UserID    string    `json:"-"`
	}

	// LoginRequest 登录提供内容
	LoginRequest struct {
		Username string `json:"username" form:"username" query:"username" `
		Password string `json:"password" form:"password" query:"password"`
	}
	//User is
	User struct {
		Username string `json:"username" form:"username" query:"username" default:"abc"`
		Password string `json:"password" form:"password" query:"password" default:"2345"`
		Level    string `json:"level" form:"level" query:"level" default:"x-xx"`
	}
)

var (
	// tokenMap = make(map[string]Token)
	users []User
)

// var mutex sync.Mutex
// fmt.Println("Lock the lock")
// mutex.Lock()
// mutex.Unlock()

// ReadUsers will
func readFileUsers(fname string) {
	bufus, err := ioutil.ReadFile(fname)
	if err != nil {
		writeFileUsers(fname)
		log.Error(err)
		return
	}

	// bufusAes, err := AesDecrypt(bufus, []byte{20, 30, 40, 99}) //password
	buf := bytes.NewBuffer(bufus)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&users); err != nil {
		log.Error(err)
		return
	}

	log.Debug("user readFileUsers ", len(users))
}

// WriteUsers will
func writeFileUsers(fname string) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(users); err != nil {
		log.Error(err)
		return
	}

	log.Debug("user writeFileUsers ", users)
	// usAes, err := AesEncrypt(buf.Bytes(), []byte{20, 30, 40, 99})
	if err := ioutil.WriteFile(fname, buf.Bytes(), 0644); err != nil {
		log.Error(err)
	}
}

// skipper 这些不需要token
func skipper(c echo.Context) bool {
	method := c.Request().Method
	path := c.Path()
	// 先处理非GET方法，除了登录，现实中还可能有一些 webhooks
	switch path {
	case
		// 登录
		"/login":
		return true
	}
	// 从这里开始必须是GET方法
	if method != "GET" {
		return false
	}
	if path == "" {
		return true
	}
	resource := strings.Split(path, "/")[1]
	switch resource {
	case
		// 公开信息，把需要公开的资源每个一行写这里
		"swagger",
		"public":
		return true
	}
	log.Debug("login skipper failure ", method, path)
	return false
}

// Validator 校验token是否合法，顺便根据token在 context中赋值 user id
func validator(token string, c echo.Context) (bool, error) {
	// 调试后门
	log.Debug("login validator token:", token)
	// if config.Debug && token == "debug" {
	// 	// c.Set("user_id", 1)
	// 	return true, nil
	// }
	// // 寻找token
	// if t, ok == tokenMap[token]; !ok {
	// 	return false, nil
	// } else if err != nil {
	// 	return false, err
	// }
	// // 设置用户
	// c.Set("user_id", t.UserID)
	// log.Debug("login validator ", tokenMap[token])

	return true, nil
}

// 这个函数还有一种设计风格，就是只是返回userid，
// 以支持可选登录，在业务中判断userid如果是0就没有登录
// func parseUser(c echo.Context) (userID int, err error) {
// 	userID, ok := c.Get("user_id").(int)
// 	if !ok || userID == 0 {
// 		return 0, ErrUnauthorized
// 	}
// 	return userID, nil
// }

// Login 登录函数，demo中为了简洁就只有user password可以通过
// 实际应用中这个函数会相当复杂，要用正则判断输入的用户名是什么类型，然后调用相关函数去找用户。
// 还要兼容第三方登录，所以请求结构体也会更加复杂。
func Login(c echo.Context) error {
	// 判断何种方式登录，小程序为提供code
	var lr = new(LoginRequest) // 输入请求
	if err := c.Bind(lr); err != nil {
		return err
	}

	for _, u := range users {
		if u.Username == lr.Username && u.Password == lr.Password {
			// 发行token
			// t := &Token{
			// 	Token:     uuid.NewV1().String(),
			// 	ExpiresAt: time.Now().Add(time.Hour * 96),
			// 	UserID:    userMap[lr.Username].Username,
			// }
			// log.Debugf("Login: lr=%v u=%v", lr, u)

			return c.JSON(http.StatusOK, "ok")
		}
	}

	return ErrAuthFailed
}

// Logout 登出，可以切换账号
func Logout(c echo.Context) error {
	var u = new(LoginRequest) // 输入请求
	if err := c.Bind(u); err != nil {
		// delete(tokenMap, u.Username)
		return c.NoContent(http.StatusNotImplemented)
	}
	return c.NoContent(http.StatusOK)
}

// GetUsers 返回所有用户账号，提供给admin管理使用。不提供用户注册，只能admin来管理。
func GetUsers(c echo.Context) error {
	log.Info("GetUsers ", len(users))
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return json.NewEncoder(c.Response()).Encode(users)
}

// CreateUser 创建增加一个账号，只能由admin完成
func CreateUser(c echo.Context) error {
	var mutex sync.Mutex
	var u = new(User) // 输入请求
	if err := c.Bind(u); err != nil {
		return err
	}

	i := -1
	for iu, uu := range users {
		if uu.Username == u.Username {
			i = iu
			break
		}
	}

	mutex.Lock()
	// userMap[u.Username] = u
	if i > -1 {
		users[i] = *u
	} else {
		users = append(users, *u)
	}
	mutex.Unlock()

	writeFileUsers(fnameUsers)
	return c.NoContent(http.StatusOK)
}

// UpdateUser 更新一个账号，只能由admin完成
func UpdateUser(c echo.Context) error {
	return CreateUser(c)
}

// DeleteUser 更新一个账号，只能由admin完成
func DeleteUser(c echo.Context) error {
	var mutex sync.Mutex
	id := c.Param("id")

	i := -1
	for iu, u := range users {
		if u.Username == id {
			i = iu
			break
		}
	}

	mutex.Lock()
	if i > -1 {
		users = append(users[:i], users[i+1:]...)
	}
	// delete(userMap, id)
	mutex.Unlock()

	writeFileUsers(fnameUsers)
	return c.NoContent(http.StatusOK)
	// return ErrNotFound
}
