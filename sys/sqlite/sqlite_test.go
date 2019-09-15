package sqlite_test

import (
	"testing"

	// Frameworks
	"github.com/djthorpe/gopi"
	sq "github.com/djthorpe/gopi-media"
	"github.com/djthorpe/gopi-media/sys/sqlite"
)

func Test_001(t *testing.T) {
	t.Log("Test_001")
}

func Test_002(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else if err := driver.Close(); err != nil {
		t.Error(err)
	}
}

func Test_003(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if sqlite, ok := driver.(sq.Connection); !ok {
			t.Error("Cannot cast connection object")
			_ = driver.(sq.Connection)
		} else {
			t.Log(sqlite)
		}
	}
}

func Test_004(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if sqlite, ok := driver.(sq.Connection); !ok {
			t.Error("Cannot cast connection object")
			_ = driver.(sq.Connection)
		} else if s, err := sqlite.Prepare("SELECT 1"); err != nil {
			t.Error(err)
		} else {
			t.Log(s)
		}
	}
}

func Test_005(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if sqlite, ok := driver.(sq.Connection); !ok {
			t.Error("Cannot cast connection object")
			_ = driver.(sq.Connection)
		} else if s, err := sqlite.Prepare("SELECT 1"); err != nil {
			t.Error(err)
		} else if result, err := sqlite.Do(s); err != nil {
			t.Error(err)
		} else {
			t.Log(result, s)
		}
	}
}

func Test_006(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if sqlite, ok := driver.(sq.Connection); !ok {
			t.Error("Cannot cast connection object")
			_ = driver.(sq.Connection)
		} else if tables, err := sqlite.Tables(); err != nil {
			t.Error(err)
		} else {
			t.Log(tables)
		}
	}
}
