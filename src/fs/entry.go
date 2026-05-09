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
}

var (
	uidCache   = make(map[int]string)
	uidCacheMu sync.RWMutex
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
		if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
			owner = lookupUser(int(stat.Uid))
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
