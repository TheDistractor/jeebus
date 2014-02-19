package jeebus

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	err := os.MkdirAll(Settings.LoggerDir, 0755)
	Check(err)

	lf := http.FileServer(http.Dir(Settings.LoggerDir))
	http.Handle("/logger/", http.StripPrefix("/logger/", lf))

	Register("io/+/+/+", &LoggerService{})
}

type LoggerService struct{ fd *os.File }

// Settings.LoggerDir is where log files get created. When the directory exists,
// the logger will store new files in it and append log items. Note that it is
// perfectly ok to create or remove this directory while the logger is running.

func (s *LoggerService) Handle(topic string, payload interface{}) {
	if !isPlainTextLine(payload.([]byte)) {
		return // filter input on if/... to only log simple plain text lines
	}

	split := strings.Split(topic, "/")
	port := split[2]
	// TODO: accepting any value right now, but non-monotonic would be a problem
	n, err := strconv.ParseInt(split[3], 10, 64)
	Check(err)
	timestamp := time.Unix(0, int64(n)*1000000)

	// automatic enabling/disabling of the logger, based on presence of dir
	_, err = os.Stat(Settings.LoggerDir)
	if err != nil {
		if s.fd != nil {
			log.Println("logger stopped")
			s.fd.Close()
			s.fd = nil
		}
		return
	}
	if s.fd == nil {
		log.Println("logger started")
	}
	// figure out name of logfile based on UTC date, with daily rotation
	datePath := dateFilename(timestamp)
	if s.fd == nil || datePath != s.fd.Name() {
		if s.fd != nil {
			s.fd.Close()
		}
		mode := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		fd, err := os.OpenFile(datePath, mode, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		s.fd = fd
	}
	// append a new log entry, here is an example of the format used:
	// 	L 01:02:03.537 usb-A40117UK OK 9 25 54 66 235 61 210 226 33 19
	hour, min, sec := timestamp.Clock()
	line := fmt.Sprintf("L %02d:%02d:%02d.%03d %s %s\n",
		hour, min, sec, timestamp.Nanosecond()/1000000, port, payload.([]byte))
	s.fd.WriteString(line)
}

func isPlainTextLine(input []byte) bool {
	if len(input) > 250 {
		return false // too long, limit is set (a bit arbitrarily) at 250 bytes
	}
	for _, b := range input {
		if b < 0x20 || b > 0x7E {
			return false // input has non-printable ASCII characters
		}
	}
	return true
}

func dateFilename(now time.Time) string {
	year, month, day := now.Date()
	path := fmt.Sprintf("%s/%d", Settings.LoggerDir, year)
	os.MkdirAll(path, os.ModePerm)
	// e.g. "./logger/2014/20140122.txt"
	return fmt.Sprintf("%s/%d.txt", path, (year*100+int(month))*100+day)
}
