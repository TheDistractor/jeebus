package opt_test

import (
	"testing"

	"github.com/jcw/jeebus"
	"github.com/jcw/jeebus/opt"
)

func TestSerial(t *testing.T) {
	expect(t, nil, nil)
}

type SerialMock struct{
	written []string
}

func (s *SerialMock) Read(b []byte) (n int, err error) {
	return //p.file.Read(b)
}

func (s *SerialMock) Write(b []byte) (n int, err error) {
	s.written = append(s.written, string(b))
	return len(b), nil
}

func TestSerialMock(t *testing.T) {
	err := jeebus.OpenDatabase()
	expect(t, err, nil)

	onceStartMessaging(t)
	
	mock := &SerialMock{}
	opt.SerialHandler("/dev/ttyUSB0", mock, "")
	
	expect(t, len(mock.written), 0)
}
