package xcfg

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type ConfigEntry struct {
	Name         string
	MajorVersion *int
	MinorVersion *int
	Value        interface{}
}

var _configEntryCache map[string]*ConfigEntry = make(map[string]*ConfigEntry)
var mutex sync.Mutex

func _addConfigEntry(name string, major *int, minor *int, val interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	key := strings.ToLower(name)
	if _, ok := _configEntryCache[key]; ok {
		ce := _configEntryCache[key]
		ce.MajorVersion = major
		ce.MinorVersion = minor
		ce.Value = val
	} else {
		_configEntryCache[key] = &ConfigEntry{name, major, minor, val}
	}
}

func init() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			if len(_configEntryCache) > 0 {
				fmt.Println("...ticker")

				mutex.Lock()
				rcfg := RemoteConfigSectionCollection{}
				rcfg.Application = GetAppName()
				rcfg.Machine = GetHostName()
				rcfg.Environment = GetEnvironment()
				rcfg.Sections = make([]*RemoteConfigSection, len(_configEntryCache))
				i := 0
				for key, val := range _configEntryCache {
					fmt.Println(key, *val.MajorVersion, *val.MinorVersion)
					rcfg.Sections[i] = &RemoteConfigSection{SectionName: strings.ToLower(key), MajorVersion: *val.MajorVersion, MinorVersion: *val.MinorVersion}
					i++
				}
				mutex.Unlock()

				rcfg_result := GetServerVersions(rcfg)
				if rcfg_result == nil || len(rcfg_result.Sections) == 0 {
					fmt.Println("...no change")
				} else {
					fmt.Println("...has change")
					cfg_folder := GetAppCfgFolder()
					for _, v := range rcfg_result.Sections {
						cfg_path := cfg_folder + "/" + _configEntryCache[v.SectionName].Name + ".config"
						sucess := DownloadRemoteCfg(v.SectionName, v.DownloadUrl, cfg_path)
						if sucess {
							major, minor := 1, 0
							is_loaded := LoadLocalCfg(cfg_path, _configEntryCache[v.SectionName].Value, &major, &minor)
							if is_loaded {
								fmt.Println(major, minor)
								_addConfigEntry(_configEntryCache[v.SectionName].Name, &major, &minor, _configEntryCache[v.SectionName].Value)
							}

						}

					}

				}
			}

		}

	}()
}
