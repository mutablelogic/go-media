package libraw_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libraw"
)

func Test_version_000(t *testing.T) {
	v := Libraw_version()
	if v == "" {
		t.Fatal("Libraw_version returned empty string")
	}
	t.Log("version=", v)
}

func Test_version_001(t *testing.T) {
	n := Libraw_versionNumber()
	if n == 0 {
		t.Fatal("Libraw_versionNumber returned 0")
	}
	t.Log("versionNumber=", n)
}

func Test_version_002(t *testing.T) {
	caps := Libraw_capabilities()
	t.Logf("capabilities=0x%x", caps)
}

func Test_version_003(t *testing.T) {
	n := Libraw_cameraCount()
	if n == 0 {
		t.Fatal("Libraw_cameraCount returned 0")
	}
	t.Log("cameraCount=", n)
}

func Test_version_004(t *testing.T) {
	list := Libraw_cameraList()
	if len(list) == 0 {
		t.Fatal("Libraw_cameraList returned empty list")
	}
	t.Logf("cameraList[0]=%q count=%d", list[0], len(list))
}

func Test_version_005(t *testing.T) {
	msg := Libraw_strerror(0)
	t.Log("strerror(0)=", msg)
}
