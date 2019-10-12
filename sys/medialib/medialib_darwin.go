/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package medialib

import (
	"os"
	"syscall"

	gopi "github.com/djthorpe/gopi"
)

func IdForFileInfo(info os.FileInfo) (int64, error) {
	if info == nil {
		return 0, gopi.ErrBadParameter
	} else if stat, ok := info.Sys().(*syscall.Stat_t); ok == false {
		return 0, gopi.ErrBadParameter
	} else {
		return int64(stat.Ino), nil
	}
}
