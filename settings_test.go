package jeebus_test

import (
	"bytes"
	"testing"

	"github.com/jcw/jeebus"
)

const fakeSettings = `
	PORT 	    = 123
	MQTT_URL    = "abc"
                
	APP_DIR     = "dir1"
	BASE_DIR    = "dir2"
	COMMON_DIR  = "dir3"
	DB_DIR      = "dir4"
	FILES_DIR   = "dir5"
	LOGGER_DIR  = "dir6"
                
	CERT_FILE   = "cert.pem"
	KEY_FILE    = "key.pem"
	
	VERBOSE_LOG = 3
	VERBOSE_RPC = true
`

func TestSettings(t *testing.T) {
	expect(t, jeebus.SettingsFound, false)

	jeebus.LoadSettings(bytes.NewBufferString(fakeSettings))

	expect(t, jeebus.Settings.Port, 123)
	expect(t, jeebus.Settings.MqttUrl, "abc")

	expect(t, jeebus.Settings.AppDir, "dir1")
	expect(t, jeebus.Settings.BaseDir, "dir2")
	expect(t, jeebus.Settings.CommonDir, "dir3")
	expect(t, jeebus.Settings.DbDir, "dir4")
	expect(t, jeebus.Settings.FilesDir, "dir5")
	expect(t, jeebus.Settings.LoggerDir, "dir6")

	expect(t, jeebus.Settings.CertFile, "cert.pem")
	expect(t, jeebus.Settings.KeyFile, "key.pem")

	expect(t, jeebus.Settings.VerboseLog, 3)
	expect(t, jeebus.Settings.VerboseRpc, true)

	jeebus.LoadSettings(bytes.NewBufferString(jeebus.DefaultSettings))
}
