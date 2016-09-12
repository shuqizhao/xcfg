package xcfg

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

func GetCurrentDirectory() string {
	if _currentDirectory == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		_currentDirectory = strings.Replace(dir, "\\", "/", -1)
	}
	return _currentDirectory
}

func GetAppCfgFolder() string {
	return GetCfgFolder() + "/" + GetAppName() + "/" + GetEnvironment()
}

func GetCfgFolder() string {
	goos := runtime.GOOS
	if goos == "darwin" {
		return "/usr/local/etc/beisen.configs"
	} else if goos == "linux" {
		return "/usr/local/etc/beisen.configs"
	} else if goos == "win" {
		return "c:\beisen.configs"
	} else {
		panic("未识别的操作系统：" + goos)
	}
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func ReadFile(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		//logger.Errorf("read file: %v error: %v", path, err)
		fmt.Printf("read file: %v error: %v", path, err)
		return nil
	}
	return b
}
