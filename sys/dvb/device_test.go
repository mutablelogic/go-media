//go:build dvb

package dvb_test

import (
	"github.com/mutablelogic/go-media/sys/dvb"
	"os"
	"testing"
)

func Test_Device_000(t *testing.T) {
	devices := dvb.Devices()
	if len(devices) == 0 {
		t.Skip("Skipping test, no devices available")
	}
	for _, device := range devices {
		t.Log(device)
	}
}

func Test_Device_001(t *testing.T) {
	devices := dvb.Devices()
	if len(devices) == 0 {
		t.Skip("Skipping test, no devices available")
	}
	for _, device := range devices {
		file, err := device.FEOpen(os.O_RDONLY)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()
		t.Log(device, " FE =>", file.Name())
	}
}
