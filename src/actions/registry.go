package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Action struct {
	Command string
	Args    []string
	Look    bool
	Browse  bool
}

type Rule struct {
	Priority int
	Match    func(name string, data []byte) bool
	Action   Action
}

type Registry struct {
	rules []Rule
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) AddMagic(offset int, magic []byte, act Action) {
	r.rules = append(r.rules, Rule{
		Priority: 0,
		Match: func(name string, data []byte) bool {
			if len(data) < offset+len(magic) {
				return false
			}
			for i, b := range magic {
				if data[offset+i] != b {
					return false
				}
			}
			return true
		},
		Action: act,
	})
}

func (r *Registry) AddExt(ext string, act Action) {
	ext = strings.ToLower(ext)
	r.rules = append(r.rules, Rule{
		Priority: 2,
		Match: func(name string, _ []byte) bool {
			return strings.HasSuffix(strings.ToLower(name), ext)
		},
		Action: act,
	})
}

func (r *Registry) AddExact(name string, act Action) {
	r.rules = append(r.rules, Rule{
		Priority: 1,
		Match: func(n string, _ []byte) bool {
			return strings.EqualFold(n, name)
		},
		Action: act,
	})
}

func (r *Registry) SetDefault(act Action) {
	r.rules = append(r.rules, Rule{
		Priority: 3,
		Match:    func(name string, data []byte) bool { return true },
		Action:   act,
	})
}

func (r *Registry) Resolve(path string) Action {
	f, err := os.Open(path)
	if err != nil {
		return Action{}
	}
	defer func() { _ = f.Close() }()

	data := make([]byte, 256)
	n, _ := f.Read(data)
	data = data[:n]

	name := filepath.Base(path)

	for _, rule := range r.rules {
		if rule.Match(name, data) {
			return rule.Action
		}
	}
	return Action{}
}

func (r *Registry) ResolveName(name string) Action {
	for _, rule := range r.rules {
		if rule.Match(name, nil) {
			return rule.Action
		}
	}
	return Action{}
}

func Execute(act Action, filePath string) error {
	if act.Look {
		fmt.Fprintf(os.Stderr, "\n[viewer] %s (not implemented yet)\n", filePath)
		return nil
	}
	cmdStr := act.Command
	for i, arg := range act.Args {
		cmdStr = strings.ReplaceAll(cmdStr, fmt.Sprintf("$%d", i+1), arg)
	}
	cmdStr = strings.ReplaceAll(cmdStr, "$F", filepath.Base(filePath))
	cmdStr = strings.ReplaceAll(cmdStr, "$P", filePath)
	cmdStr = strings.ReplaceAll(cmdStr, "$D", filepath.Dir(filePath))

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		b := buf[i]
		if b == 0 {
			return true
		}
	}
	return false
}

var extLangMap = map[string]string{
	".go":  "go",
	".rs":  "rust",
	".py":  "python",
	".js":  "javascript",
	".ts":  "typescript",
	".rb":  "ruby",
	".swift": "swift",
	".c":   "c",
	".h":   "c",
	".cpp": "cpp",
	".hpp": "cpp",
	".java": "java",
	".sh":  "bash",
	".zsh": "bash",
	".bash": "bash",
	".yaml": "yaml",
	".yml": "yaml",
	".json": "json",
	".xml": "xml",
	".md":  "markdown",
}

var magicPatterns = []struct {
	offset int
	pat    string
	act    Action
}{
	{0, "\x89PNG", Action{Look: true}},
	{0, "\xff\xd8\xff", Action{Look: true}},
	{0, "GIF8", Action{Look: true}},
}

var extActions = map[string]Action{
	".txt":  {Look: true},
	".md":   {Look: true},
	".go":   {Look: true},
	".rs":   {Look: true},
	".py":   {Look: true},
	".js":   {Look: true},
	".ts":   {Look: true},
	".c":    {Look: true},
	".h":    {Look: true},
	".cpp":  {Look: true},
	".java": {Look: true},
	".sh":   {Look: true},
	".bash": {Look: true},
	".zsh":  {Look: true},
	".yaml": {Look: true},
	".yml":  {Look: true},
	".json": {Look: true},
	".xml":  {Look: true},
	".csv":  {Look: true},
	".toml": {Look: true},
	".conf": {Look: true},
	".log":  {Look: true},
	".png":  {Command: "open $P"},
	".jpg":  {Command: "open $P"},
	".jpeg": {Command: "open $P"},
	".gif":  {Command: "open $P"},
	".bmp":  {Command: "open $P"},
	".webp": {Command: "open $P"},
	".mdx":  {Command: "open -a MP4M.app $P"},
	".zip":  {Browse: true},
	".tar":  {Browse: true},
	".tgz":  {Browse: true},
	".gz":   {Browse: true},
	".bz2":  {Browse: true},
	".7z":   {Browse: true},
}

func DefaultRegistry() *Registry {
	r := NewRegistry()

	for _, m := range magicPatterns {
		pat := []byte(regexp.QuoteMeta(m.pat))
		r.AddMagic(m.offset, pat, m.act)
	}

	ext := extActions
	ext[".go"] = Action{Look: true}

	for ext_, act := range ext {
		r.AddExt(ext_, act)
	}

	return r
}

func DetectLang(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if lang, ok := extLangMap[ext]; ok {
		return lang
	}
	return ""
}
