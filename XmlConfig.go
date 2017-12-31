package xcfg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type RemoteConfigSectionCollection struct {
	Machine     string                 `xml:"machine,attr"`
	Application string                 `xml:"application,attr"`
	Environment string                 `xml:"env,attr"`
	Sections    []*RemoteConfigSection `xml:"section"`
}

func (rcfg RemoteConfigSectionCollection) Count() int {
	if rcfg.Sections == nil {
		return 0
	} else {
		return len(rcfg.Sections)
	}
}

type RemoteConfigSection struct {
	SectionName  string `xml:"name,attr"`
	MajorVersion int    `xml:"majorVerion,attr"`
	MinorVersion int    `xml:"minorVerion,attr"`
	DownloadUrl  string `xml:"downloadUrl,attr"`
	TemplateUrl  string `xml:"templateUrl,attr"`
}

func LoadLocalCfg(path string, v interface{}, major *int, minor *int) bool {
	if !Exist(path) {
		return false
	}
	b := ReadFile(path)
	if v != nil {
		err := xml.Unmarshal(b, v)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	b_str := string(b)
	decoder := xml.NewDecoder(strings.NewReader(b_str))
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			//name := token.Name.Local
			//fmt.Printf("Token name: %s\n", name)
			for _, attr := range token.Attr {
				attrName := attr.Name.Local
				attrValue := attr.Value
				if attrName == "majorVersion" {
					*major, _ = strconv.Atoi(attrValue)
				}
				if attrName == "minorVersion" {
					*minor, _ = strconv.Atoi(attrValue)
				}
				//fmt.Printf("An attribute is: %s %s\n", attrName, attrValue)
			}
		default:
			// ...
		}
	}
	return true
}

func GetRemoteConfigSectionParam(cfgName string,major int,minor int) *RemoteConfigSection {
	rcfg := RemoteConfigSectionCollection{}
	rcfg.Application = GetAppName()
	rcfg.Machine = GetHostName()
	rcfg.Environment = GetEnvironment()
	rcfg.Sections = make([]*RemoteConfigSection, 1)
	rcfg.Sections[0] = &RemoteConfigSection{SectionName: strings.ToLower(cfgName), MajorVersion: major, MinorVersion: minor}
	rcfg_result := GetServerVersions(rcfg)
	if rcfg_result == nil || len(rcfg_result.Sections) == 0 {
		return nil
	} else {
		return rcfg_result.Sections[0]
	}
}

func GetServerVersions(rcfg RemoteConfigSectionCollection) *RemoteConfigSectionCollection {

	b, _ := xml.Marshal(rcfg)
	//fmt.Println(string(b))
	body := bytes.NewBuffer(b)
	//fmt.Println(GetRemoteCfgUrl())

	client := &http.Client{}

	request, _ := http.NewRequest("POST", GetRemoteCfgUrl(), body)

	request.Header.Add("Content-Type", "text/xml")
	resp, err := client.Do(request)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		//TODO EROR
		fmt.Println(err)
		return nil
	} else {
		rcfg_result := RemoteConfigSectionCollection{}
		data, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(data))
		xml.Unmarshal(data, &rcfg_result)
		return &rcfg_result
	}
}

func DownloadRemoteCfg(sectionName string, url string, targetPath string) bool {
	if !strings.HasPrefix(url, "http") {
		url = GetRemoteCfgShortUrl() + "/" + url
		//fmt.Println(url)
	}
	resp, err := http.Get(url)

	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		//TODO EROR
		fmt.Println(err)
		return false
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(data))
		if len(data)==0 {
			return false
		}
		tmpFile := targetPath + "." + GetGuid()
		fs, _ := os.Create(tmpFile)
		fs.Write(data)
		fs.Close()
		if Exist(targetPath) {
			os.Remove(targetPath)
		}
		os.Rename(tmpFile, targetPath)
		os.Remove(tmpFile)
		return true
	}
}

func LoadCfg(entity interface{}) {
	entity_type := reflect.TypeOf(entity)
	entity_elem := entity_type.Elem()

	entity_value := reflect.ValueOf(entity)
	entity_indirect := reflect.Indirect(entity_value)

	cfg_name := entity_indirect.Type().Name()

	for i := 0; i < entity_elem.NumField(); i++ {
		field := entity_elem.Field(i)
		if field.Type.PkgPath() == "encoding/xml" && field.Type.Name() == "Name" {
			cfg_name = field.Tag.Get("xml")
			break
		}
	}

	cfg_folder := GetAppCfgFolder()
	if !Exist(cfg_folder) {
		err := os.MkdirAll(cfg_folder, 0777)
		if err != nil {
			panic(err)
		}
	}
	cfg_path := cfg_folder + "/" + cfg_name + ".config"

	//fmt.Println(cfg_path)

	major, minor := 1, 0
	is_loaded := LoadLocalCfg(cfg_path, entity, &major, &minor)
	if is_loaded {
		fmt.Println(major, minor)
		if isFirstLoad(cfg_name){
			_addConfigEntry(cfg_name, &major, &minor, entity)
		}else {
			_addConfigEntry(cfg_name, &major, &minor, entity)
			return
		}
	}
	param := GetRemoteConfigSectionParam(cfg_name,major,minor)
	if param != nil {
		//fmt.Println(param.DownloadUrl)
		sucess := DownloadRemoteCfg(cfg_name, param.DownloadUrl, cfg_path)
		if sucess {
			is_loaded = LoadLocalCfg(cfg_path, entity, &major, &minor)
			if is_loaded {
				_addConfigEntry(cfg_name, &major, &minor, entity)
			}
		}
		template_cfg_path := cfg_folder + "/" + cfg_name + ".template"
		DownloadRemoteCfg(cfg_name, param.TemplateUrl, template_cfg_path)
	}

}
