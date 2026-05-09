package fs

import (
	"sort"
	"strings"
)

type DiffKind int

const (
	DiffSame DiffKind = iota
	DiffLeftOnly
	DiffRightOnly
	DiffDifferent
)

type DiffEntry struct {
	Name      string
	Kind      DiffKind
	LeftSize  int64
	RightSize int64
}

func CompareDirs(left, right *Listing) []DiffEntry {
	leftMap := make(map[string]Entry, len(left.Entries))
	for _, e := range left.Entries {
		leftMap[e.Name] = e
	}
	rightMap := make(map[string]Entry, len(right.Entries))
	for _, e := range right.Entries {
		rightMap[e.Name] = e
	}

	allNames := make(map[string]bool)
	for _, e := range left.Entries {
		allNames[e.Name] = true
	}
	for _, e := range right.Entries {
		allNames[e.Name] = true
	}

	var result []DiffEntry
	for name := range allNames {
		le, hasLeft := leftMap[name]
		re, hasRight := rightMap[name]
		switch {
		case hasLeft && !hasRight:
			result = append(result, DiffEntry{Name: name, Kind: DiffLeftOnly, LeftSize: le.Size})
		case !hasLeft && hasRight:
			result = append(result, DiffEntry{Name: name, Kind: DiffRightOnly, RightSize: re.Size})
		case le.Size != re.Size || !le.ModTime.Equal(re.ModTime):
			result = append(result, DiffEntry{
				Name: name, Kind: DiffDifferent,
				LeftSize: le.Size, RightSize: re.Size,
			})
		default:
			result = append(result, DiffEntry{Name: name, Kind: DiffSame})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	return result
}
