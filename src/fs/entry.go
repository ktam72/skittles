package fs

import (
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type SortMode int

const (
	SortName SortMode = iota
	SortTime
	SortExt
	SortSize
)

type Entry struct {
	Name     string
	Path     string
	Size     int64
	Mode     os.FileMode
	ModTime  time.Time
	IsDir    bool
	IsLink   bool
	IsMarked bool
	Owner    string
	Group    string
}

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

type Listing struct {
	Entries []Entry
	SortBy  SortMode
	Reverse bool
}

func ReadDir(dir string) (*Listing, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	l := &Listing{SortBy: SortName}

	parent := filepath.Dir(strings.TrimRight(dir, "/"))
	if dir != "/" {
		l.Entries = append(l.Entries, Entry{
			Name:   "..",
			Path:   parent,
			Size:   0,
			Mode:   os.ModeDir | 0755,
			IsDir:  true,
			IsLink: false,
		})
	}

	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			continue
		}
		isLink := fi.Mode()&os.ModeSymlink != 0
		if isLink {
			if target, err := os.Stat(filepath.Join(dir, e.Name())); err == nil {
				fi = target
			}
		}
		owner := ""
		group := ""
		if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
			owner = lookupUser(int(stat.Uid))
			group = lookupGroup(int(stat.Gid))
		}
		l.Entries = append(l.Entries, Entry{
			Name:    e.Name(),
			Path:    filepath.Join(dir, e.Name()),
			Size:    fi.Size(),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
			IsLink:  isLink,
			Owner:   owner,
			Group:   group,
		})
	}
	l.Sort()
	return l, nil
}

func (l *Listing) Sort() {
	sort.SliceStable(l.Entries, func(i, j int) bool {
		a, b := &l.Entries[i], &l.Entries[j]
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		var less bool
		switch l.SortBy {
		case SortName:
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case SortTime:
			less = a.ModTime.Before(b.ModTime)
		case SortExt:
			less = filepath.Ext(a.Name) < filepath.Ext(b.Name)
		case SortSize:
			less = a.Size < b.Size
		}
		if l.Reverse {
			return !less
		}
		return less
	})
}
