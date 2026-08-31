//go:build unix

package fs

import (
	"os"
	"os/user"
	"strconv"
	"sync"
	"syscall"
)

var (
	uidCache   = make(map[int]string)
	uidCacheMu sync.RWMutex
	gidCache   = make(map[int]string)
	gidCacheMu sync.RWMutex
)

func lookupUser(uid int) string {
	uidCacheMu.RLock()
	s, ok := uidCache[uid]
	uidCacheMu.RUnlock()
	if ok {
		return s
	}
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		s = strconv.Itoa(uid)
	} else {
		s = u.Username
	}
	uidCacheMu.Lock()
	uidCache[uid] = s
	uidCacheMu.Unlock()
	return s
}

func lookupGroup(gid int) string {
	gidCacheMu.RLock()
	s, ok := gidCache[gid]
	gidCacheMu.RUnlock()
	if ok {
		return s
	}
	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		s = strconv.Itoa(gid)
	} else {
		s = g.Name
	}
	gidCacheMu.Lock()
	gidCache[gid] = s
	gidCacheMu.Unlock()
	return s
}

func lookupStat(fi os.FileInfo) (string, string) {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		return lookupUser(int(stat.Uid)), lookupGroup(int(stat.Gid))
	}
	return "", ""
}
