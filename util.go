package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"log"
	"net"
)

// FileURL 图片链接
type FileURL string

// // ToString 转换为string类型
// func (f FileURL) ToString() string {
// 	var s = string(f)
// 	var url = s
// 	if !strings.HasPrefix(s, "http") {
// 		url = config.FileURL + s
// 	}
// 	return url
// }

// // MarshalJSON 转换为json类型 加域名
// func (f FileURL) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(f.ToString())
// }

// // UnmarshalJSON 不做处理
// func (f *FileURL) UnmarshalJSON(data []byte) error {
// 	var tmp string
// 	if err := json.Unmarshal(data, &tmp); err != nil {
// 		return err
// 	}
// 	tmp = strings.TrimPrefix(tmp, config.FileURL)
// 	*f = FileURL(tmp)
// 	return nil
// }

// // Scan implements the Scanner interface.
// func (f *FileURL) Scan(src interface{}) error {
// 	if src == nil {
// 		*f = ""
// 		return nil
// 	}
// 	tmp, ok := src.([]byte)
// 	if !ok {
// 		return errors.New("Read file url data from DB failed")
// 	}
// 	*f = FileURL(tmp)
// 	return nil
// }

// Value implements the driver Valuer interface.
// func (f FileURL) Value() (driver.Value, error) {
// 	return string(f), nil
// }

// GetLocalhostIP is read local host info
func GetLocalhostIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

//PKCS5Padding will
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//PKCS5UnPadding will
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//ZeroPadding will
func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

//ZeroUnPadding will
func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AesEncrypt will
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AesDecrypt will
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
