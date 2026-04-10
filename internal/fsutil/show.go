package fsutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Entry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
	SizeBytes int64     `json:"size_bytes"`
	Executable bool     `json:"executable"`
	Mode      string    `json:"mode"`
	Modified  time.Time `json:"modified"`
	GitStatus string    `json:"git_status,omitempty"`
	Depth     int       `json:"depth,omitempty"`
}

type Options struct {
	Path      string
	SortBy    string
	Reverse   bool
	Hidden    bool
	Type      string
	Ext       []string
	Largest   int
	Recent    int
	Tree      bool
	Depth     int
	GitStatus bool
}

func Show(opts Options) ([]Entry, error) {
	path := opts.Path
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if opts.Tree {
		entries, err = walkEntries(absPath, opts.Depth)
	} else {
		entries, err = listEntries(absPath)
	}
	if err != nil {
		return nil, err
	}

	entries = filterEntries(entries, opts)
	sortEntries(entries, opts.SortBy, opts.Reverse)
	entries = cutEntries(entries, opts)

	if opts.GitStatus {
		statuses := gitStatuses(absPath)
		for i := range entries {
			if s, ok := statuses[filepath.Clean(entries[i].Path)]; ok {
				entries[i].GitStatus = s
			}
		}
	}
	return entries, nil
}

func listEntries(path string) ([]Entry, error) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	res := make([]Entry, 0, len(dirs)+2)

	if info, err := os.Lstat(path); err == nil {
		res = append(res, toEntry(path, ".", info, 0))
	}
	parent := filepath.Dir(path)
	if info, err := os.Lstat(parent); err == nil {
		res = append(res, toEntry(parent, "..", info, 0))
	}

	for _, d := range dirs {
		full := filepath.Join(path, d.Name())
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		res = append(res, toEntry(full, d.Name(), info, 0))
	}
	return res, nil
}

func walkEntries(path string, maxDepth int) ([]Entry, error) {
	res := make([]Entry, 0)
	rootDepth := strings.Count(filepath.Clean(path), string(os.PathSeparator))
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if p == path {
			return nil
		}
		depth := strings.Count(filepath.Clean(p), string(os.PathSeparator)) - rootDepth
		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := os.Lstat(p)
		if err != nil {
			return nil
		}
		res = append(res, toEntry(p, filepath.Base(p), info, depth))
		return nil
	})
	return res, err
}

func toEntry(fullPath, name string, info os.FileInfo, depth int) Entry {
	typeName := "file"
	if info.Mode()&os.ModeSymlink != 0 {
		typeName = "link"
	} else if info.IsDir() {
		typeName = "dir"
	}
	return Entry{
		Name:      name,
		Path:      fullPath,
		Type:      typeName,
		SizeBytes: info.Size(),
		Executable: info.Mode().Perm()&0o111 != 0 && !info.IsDir(),
		Mode:      info.Mode().String(),
		Modified:  info.ModTime(),
		Depth:     depth,
	}
}

func filterEntries(entries []Entry, opts Options) []Entry {
	extSet := map[string]struct{}{}
	for _, x := range opts.Ext {
		x = strings.TrimSpace(strings.ToLower(x))
		if x == "" {
			continue
		}
		if !strings.HasPrefix(x, ".") {
			x = "." + x
		}
		extSet[x] = struct{}{}
	}

	res := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if !opts.Hidden && strings.HasPrefix(e.Name, ".") && e.Name != "." && e.Name != ".." {
			continue
		}
		if opts.Type != "" && opts.Type != "all" && e.Type != opts.Type {
			continue
		}
		if len(extSet) > 0 && e.Type == "file" {
			ext := strings.ToLower(filepath.Ext(e.Name))
			if _, ok := extSet[ext]; !ok {
				continue
			}
		}
		res = append(res, e)
	}
	return res
}

func sortEntries(entries []Entry, sortBy string, reverse bool) {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy == "" {
		sortBy = "name"
	}

	less := func(i, j int) bool {
		if specialRank(entries[i].Name) != specialRank(entries[j].Name) {
			return specialRank(entries[i].Name) < specialRank(entries[j].Name)
		}
		if entries[i].Type != entries[j].Type {
			return entries[i].Type == "dir"
		}
		switch sortBy {
		case "size":
			if entries[i].SizeBytes == entries[j].SizeBytes {
				return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			}
			return entries[i].SizeBytes < entries[j].SizeBytes
		case "time":
			if entries[i].Modified.Equal(entries[j].Modified) {
				return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			}
			return entries[i].Modified.Before(entries[j].Modified)
		default:
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if reverse {
			return !less(i, j)
		}
		return less(i, j)
	})
}

func specialRank(name string) int {
	switch name {
	case ".":
		return 0
	case "..":
		return 1
	default:
		return 2
	}
}

func cutEntries(entries []Entry, opts Options) []Entry {
	if opts.Largest > 0 {
		sort.Slice(entries, func(i, j int) bool { return entries[i].SizeBytes > entries[j].SizeBytes })
		if len(entries) > opts.Largest {
			entries = entries[:opts.Largest]
		}
		return entries
	}
	if opts.Recent > 0 {
		sort.Slice(entries, func(i, j int) bool { return entries[i].Modified.After(entries[j].Modified) })
		if len(entries) > opts.Recent {
			entries = entries[:opts.Recent]
		}
	}
	return entries
}

func gitStatuses(path string) map[string]string {
	res := map[string]string{}
	topRaw, err := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return res
	}
	repo := strings.TrimSpace(string(topRaw))
	if repo == "" {
		return res
	}

	out, err := exec.Command("git", "-C", repo, "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		return res
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(strings.TrimSpace(line)) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		target := strings.TrimSpace(line[3:])
		if strings.Contains(target, " -> ") {
			parts := strings.Split(target, " -> ")
			target = parts[len(parts)-1]
		}
		abs := filepath.Clean(filepath.Join(repo, target))
		res[abs] = status
	}
	return res
}

func HumanSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(size)
	u := 0
	for f >= 1024 && u < len(units)-1 {
		f /= 1024
		u++
	}
	if u == 0 {
		return fmt.Sprintf("%d%s", int64(f), units[u])
	}
	return fmt.Sprintf("%.1f%s", f, units[u])
}
