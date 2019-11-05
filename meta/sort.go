package meta

import (
	"time"

	conf "github.com/solozyx/object-storage/config"
)

type ByUploadTime []FileMeta

func (a ByUploadTime) Len() int {
	return len(a)
}

func (a ByUploadTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByUploadTime) Less(i, j int) bool {
	iTime, _ := time.Parse(conf.SysTimeform, a[i].UploadAt)
	jTime, _ := time.Parse(conf.SysTimeform, a[j].UploadAt)
	return iTime.UnixNano() < jTime.UnixNano()
}
