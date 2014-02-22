package jeebus

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
)

const DefaultSettings = `
# the settings in this file are of the form: <name> "=" <json-value>

# HTTP and MQTT server settings
PORT 	 = 3000
MQTT_URL = "tcp://0.0.0.0:1883"

# directory path settings
APP_DIR    = "./app"
BASE_DIR   = "./base"
COMMON_DIR = "./common"
DB_DIR     = "./db"
FILES_DIR  = "./files"
LOGGER_DIR = "./logger"
`

var Settings = struct {
	Port    int    `json:",string"`
	MqttUrl string `json:",string"`

	AppDir    string `json:",string"`
	BaseDir   string `json:",string"`
	CommonDir string `json:",string"`
	DbDir     string `json:",string"`
	FilesDir  string `json:",string"`
	LoggerDir string `json:",string"`

	CertFile string `json:",string"`
	KeyFile  string `json:",string"`
}{}

var SettingsFound = initSettings() // nasty side-effect is to do a "pre-init"

func initSettings() bool {
	LoadSettings(bytes.NewBufferString(DefaultSettings))

	filename := os.Getenv("SETTINGS")
	if filename == "" {
		filename = "settings.txt"
	}

	fd, err := os.Open(filename)
	if err == nil {
		defer fd.Close()
		LoadSettings(fd)
	}

	return err == nil
}

func LoadSettings(fd io.Reader) {
	smap := make(map[string]string)

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t")
		if line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.SplitN(line, "=", 2)
			if len(fields) != 2 {
				log.Fatalln("cannot parse settings:", scanner.Text())
			}
			key := strings.Trim(fields[0], " \t")
			value := strings.Trim(fields[1], " \t")
			env := os.Getenv(key)
			if env != "" {
				value = env
			}
			smap[CapsToIdentifier(key)] = value
		}
	}

	err := json.Unmarshal([]byte(ToJson(smap)), &Settings)
	Check(err)
}

func CapsToIdentifier(s string) (result string) {
	s = strings.ToLower(s)
	t := strings.Split(s, "_")
	for i, _ := range t {
		result += strings.ToUpper(t[i][:1]) + t[i][1:]
	}
	return
}
