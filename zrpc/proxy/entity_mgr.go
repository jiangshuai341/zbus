package proxy

import "sync"

type entityMgr struct {
	sync.Map
}
