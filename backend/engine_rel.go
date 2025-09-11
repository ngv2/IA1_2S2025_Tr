package main

import (
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"

	golog "github.com/mndrix/golog"
)

var (
	plFacts2 = map[string]map[[2]string]bool{}
	plFacts3 = map[string]map[[3]string]bool{}
)

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func normalizeNumber(s string) (string, bool) {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", false
	}
	return strconv.FormatFloat(v, 'g', -1, 64), true
}

func plBuildProgram2and3() string {
	var b strings.Builder
	for pred, set := range plFacts2 {
		for pair := range set {
			b.WriteString(pred + "(" + pair[0] + "," + pair[1] + ").\n")
		}
	}
	for pred, set := range plFacts3 {
		for tri := range set {
			b.WriteString(pred + "(" + tri[0] + "," + tri[1] + "," + tri[2] + ").\n")
		}
	}
	return b.String()
}

func plRebuildMachineAll() {
	code := plBuildProgram() + plBuildProgram2and3()
	plMachine = golog.NewMachine().Consult(code)
}

func plSavePred(file, pred string, newLines []string) error {
	var keep []string
	if old, err := os.ReadFile(file); err == nil {
		lines := strings.Split(string(old), "\n")
		prefix := pred + "("
		for _, ln := range lines {
			t := strings.TrimSpace(ln)
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, prefix) && strings.HasSuffix(t, ").") {
				continue
			}
			keep = append(keep, ln)
		}
	}
	var b strings.Builder
	for _, ln := range keep {
		b.WriteString(ln)
		if !strings.HasSuffix(ln, "\n") {
			b.WriteByte('\n')
		}
	}
	for _, ln := range newLines {
		b.WriteString(ln)
		if !strings.HasSuffix(ln, "\n") {
			b.WriteByte('\n')
		}
	}
	return os.WriteFile(file, []byte(b.String()), fs.FileMode(0644))
}

func plSave2(pred string) error {
	file := plFiles[pred]
	if file == "" {
		return nil
	}
	var lines []string
	for pair := range plFacts2[pred] {
		lines = append(lines, pred+"("+pair[0]+","+pair[1]+").")
	}
	return plSavePred(file, pred, lines)
}

func plSave3(pred string) error {
	file := plFiles[pred]
	if file == "" {
		return nil
	}
	var lines []string
	for tri := range plFacts3[pred] {
		lines = append(lines, pred+"("+tri[0]+","+tri[1]+","+tri[2]+").")
	}
	return plSavePred(file, pred, lines)
}

func Register2(pred, file string) error {
	plMutex.Lock()
	defer plMutex.Unlock()
	if _, ok := plFacts2[pred]; !ok {
		plFacts2[pred] = map[[2]string]bool{}
	}
	plFiles[pred] = file
	if src, err := os.ReadFile(file); err == nil {
		rx := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(pred) + `\(([^,]+),\s*([^,)]+)\)\.\s*$`)
		for _, ln := range strings.Split(string(src), "\n") {
			if m := rx.FindStringSubmatch(ln); m != nil {
				a := toAtom(trimQuotes(m[1]))
				b := toAtom(trimQuotes(m[2]))
				plFacts2[pred][[2]string{a, b}] = true
			}
		}
	}
	plRebuildMachineAll()
	return nil
}

func Register3(pred, file string) error {
	plMutex.Lock()
	defer plMutex.Unlock()
	if _, ok := plFacts3[pred]; !ok {
		plFacts3[pred] = map[[3]string]bool{}
	}
	plFiles[pred] = file
	if src, err := os.ReadFile(file); err == nil {
		rx := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(pred) + `\(([^,]+),\s*([^,]+),\s*([^,)]+)\)\.\s*$`)
		for _, ln := range strings.Split(string(src), "\n") {
			if m := rx.FindStringSubmatch(ln); m != nil {
				a := toAtom(trimQuotes(m[1]))
				b := toAtom(trimQuotes(m[2]))
				if w, ok := normalizeNumber(m[3]); ok {
					plFacts3[pred][[3]string{a, b, w}] = true
				}
			}
		}
	}
	plRebuildMachineAll()
	return nil
}

func List2(pred string) [][2]string {
	plMutex.Lock()
	defer plMutex.Unlock()
	set := plFacts2[pred]
	out := make([][2]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	return out
}

func List3(pred string) [][3]string {
	plMutex.Lock()
	defer plMutex.Unlock()
	set := plFacts3[pred]
	out := make([][3]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	return out
}

func Create2(pred, aRaw, bRaw string) ([2]string, bool) {
	plMutex.Lock()
	defer plMutex.Unlock()
	a := toAtom(aRaw)
	b := toAtom(bRaw)
	if _, ok := plFacts2[pred]; !ok {
		plFacts2[pred] = map[[2]string]bool{}
	}
	key := [2]string{a, b}
	if plFacts2[pred][key] {
		return key, false
	}
	plFacts2[pred][key] = true
	plRebuildMachineAll()
	_ = plSave2(pred)
	return key, true
}

func Create3(pred, aRaw, bRaw, wRaw string) ([3]string, bool) {
	plMutex.Lock()
	defer plMutex.Unlock()
	a := toAtom(aRaw)
	b := toAtom(bRaw)
	w, ok := normalizeNumber(wRaw)
	if !ok {
		return [3]string{}, false
	}
	if _, ok := plFacts3[pred]; !ok {
		plFacts3[pred] = map[[3]string]bool{}
	}
	key := [3]string{a, b, w}
	if plFacts3[pred][key] {
		return key, false
	}
	plFacts3[pred][key] = true
	plRebuildMachineAll()
	_ = plSave3(pred)
	return key, true
}

func Delete2(pred, aRaw, bRaw string) bool {
	plMutex.Lock()
	defer plMutex.Unlock()
	a := toAtom(aRaw)
	b := toAtom(bRaw)
	set, ok := plFacts2[pred]
	if !ok {
		return false
	}
	key := [2]string{a, b}
	if !set[key] {
		return false
	}
	delete(set, key)
	plRebuildMachineAll()
	_ = plSave2(pred)
	return true
}

func Delete3(pred, aRaw, bRaw, wRaw string) bool {
	plMutex.Lock()
	defer plMutex.Unlock()
	a := toAtom(aRaw)
	b := toAtom(bRaw)
	w, ok := normalizeNumber(wRaw)
	if !ok {
		return false
	}
	set, ok2 := plFacts3[pred]
	if !ok2 {
		return false
	}
	key := [3]string{a, b, w}
	if !set[key] {
		return false
	}
	delete(set, key)
	plRebuildMachineAll()
	_ = plSave3(pred)
	return true
}

func Update2(pred, oldA, oldB, newA, newB string) ([2]string, bool, string) {
	plMutex.Lock()
	defer plMutex.Unlock()
	o := [2]string{toAtom(oldA), toAtom(oldB)}
	n := [2]string{toAtom(newA), toAtom(newB)}
	set, ok := plFacts2[pred]
	if !ok || !set[o] {
		return [2]string{}, false, "not_found"
	}
	if o == n {
		return n, true, ""
	}
	if set[n] {
		return [2]string{}, false, "conflict"
	}
	delete(set, o)
	set[n] = true
	plRebuildMachineAll()
	_ = plSave2(pred)
	return n, true, ""
}

func Update3(pred, oldA, oldB, oldW, newA, newB, newW string) ([3]string, bool, string) {
	plMutex.Lock()
	defer plMutex.Unlock()
	ow, ok1 := normalizeNumber(oldW)
	nw, ok2 := normalizeNumber(newW)
	if !ok1 || !ok2 {
		return [3]string{}, false, "bad_number"
	}
	o := [3]string{toAtom(oldA), toAtom(oldB), ow}
	n := [3]string{toAtom(newA), toAtom(newB), nw}
	set, ok := plFacts3[pred]
	if !ok || !set[o] {
		return [3]string{}, false, "not_found"
	}
	if o == n {
		return n, true, ""
	}
	if set[n] {
		return [3]string{}, false, "conflict"
	}
	delete(set, o)
	set[n] = true
	plRebuildMachineAll()
	_ = plSave3(pred)
	return n, true, ""
}
