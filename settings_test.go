package jeebus_test

import (
	"bytes"
	"testing"

	"github.com/jcw/jeebus"
)

const fakeSettings = `
	PORT 	   = 123
	MQTT_URL   = "abc"

	APP_DIR    = "dir1"
	COMMON_DIR = "dir2"
	DB_DIR     = "dir3"
	FILES_DIR  = "dir4"
	LOGGER_DIR = "dir5"

	CERT_FILE  = "cert.pem"
	KEY_FILE   = "key.pem"
`

func TestSettings(t *testing.T) {
	expect(t, jeebus.SettingsFound, false)

	jeebus.LoadSettings(bytes.NewBufferString(fakeSettings))

	expect(t, jeebus.Settings.Port, 123)
	expect(t, jeebus.Settings.MqttUrl, "abc")

	expect(t, jeebus.Settings.AppDir, "dir1")
	expect(t, jeebus.Settings.CommonDir, "dir2")
	expect(t, jeebus.Settings.DbDir, "dir3")
	expect(t, jeebus.Settings.FilesDir, "dir4")
	expect(t, jeebus.Settings.LoggerDir, "dir5")

	expect(t, jeebus.Settings.CertFile, "cert.pem")
	expect(t, jeebus.Settings.KeyFile, "key.pem")

	jeebus.LoadSettings(bytes.NewBufferString(jeebus.DefaultSettings))
}
