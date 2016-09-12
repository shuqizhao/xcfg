package xcfg

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const middle = "========="

const (
	Dev     = "dev"
	Testing = "testing"
	Labs    = "labs"
	Product = "prod"
)

var (
	_currentDirectory string
	_appCfgInstance   *MiniConfig
	_appName          string
	_environment      string
	_app_cfg_name     = "app.cfg"
	_remote_cfg_url   string
)

type MiniConfig struct {
	Mymap  map[string]string
	strcet string
}

func (c *MiniConfig) InitConfig(path string) {
	c.Mymap = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		s := strings.TrimSpace(string(b))
		//fmt.Println(s)
		if strings.Index(s, "#") == 0 {
			continue
		}

		n1 := strings.Index(s, "[")
		n2 := strings.LastIndex(s, "]")
		if n1 > -1 && n2 > -1 && n2 > n1+1 {
			c.strcet = strings.TrimSpace(s[n1+1 : n2])
			continue
		}

		if len(c.strcet) == 0 {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		frist := strings.TrimSpace(s[:index])
		if len(frist) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])

		pos := strings.Index(second, "\t#")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " #")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, "\t//")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " //")
		if pos > -1 {
			second = second[0:pos]
		}

		if len(second) == 0 {
			continue
		}

		key := c.strcet + middle + frist
		c.Mymap[key] = strings.TrimSpace(second)
	}
}

func (c *MiniConfig) Read(node, key string) string {
	key = node + middle + key
	v, found := c.Mymap[key]
	if !found {
		return ""
	}
	return v
}

func (c *MiniConfig) Get(key string) string {
	environment := GetEnvironment()
	return c.Read(environment, key)
}

func AppCfgInstance() *MiniConfig {
	if _appCfgInstance == nil {
		_appCfgInstance = new(MiniConfig)
		_appCfgInstance.InitConfig(GetCurrentDirectory() + "/" + _app_cfg_name)
	}
	return _appCfgInstance
}

func GetAppName() string {
	if _appName == "" {
		_appName = AppCfgInstance().Get("app_name")
	}
	return _appName
}

func GetEnvironment() string {
	if _environment == "" {
		ei := AppCfgInstance().Read("default", "environment")
		dlei := strings.ToLower(strings.Trim(ei, ""))
		if dlei == Product {
			_environment = Product
		} else if dlei == Labs {
			_environment = Labs
		} else if dlei == Testing {
			_environment = Testing
		} else {
			_environment = Dev
		}
	}
	return _environment
}

func GetRemoteCfgUrl() string {
	if _remote_cfg_url == "" {
		host := AppCfgInstance().Get("remote_cfg_host")
		deal_host := strings.ToLower(strings.Trim(host, ""))

		port := AppCfgInstance().Get("remote_cfg_port")
		deal_port := strings.ToLower(strings.Trim(port, ""))

		_remote_cfg_url = "http://" + deal_host + ":" + deal_port + "/ConfigVersionHandler.ashx"
	}
	return _remote_cfg_url
}

func GetHostName() string {
	host, err := os.Hostname()
	if err == nil {

	}
	return host
}
