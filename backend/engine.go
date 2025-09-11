package main

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"

	golog "github.com/mndrix/golog"
)

var (
	plMutex    sync.Mutex
	plMachine  golog.Machine
	plFacts    = map[string]map[string]bool{} // predicado -> set de átomos
	plFiles    = map[string]string{}          // predicado -> archivo
)

func toAtom(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^\p{L}\p{N}_]+`)
	s = re.ReplaceAllString(s, "_")
	if s == "" {
		s = "x"
	}
	return s
}

func plBuildProgram() string {
	var b strings.Builder
	for pred, set := range plFacts {
		for id := range set {
			b.WriteString(pred)
			b.WriteString("(")
			b.WriteString(id)
			b.WriteString(").\n")
		}
	}
	return b.String()
}

func plRebuildMachine() {
	code := plBuildProgram()
	plMachine = golog.NewMachine().Consult(code)
}

func plSave(pred string) error {
	file, ok := plFiles[pred]
	if !ok || file == "" {
		return nil
	}
	var b strings.Builder
	for id := range plFacts[pred] {
		b.WriteString(pred)
		b.WriteString("(")
		b.WriteString(id)
		b.WriteString(").\n")
	}
	return os.WriteFile(file, []byte(b.String()), fs.FileMode(0644))
}

func PLRegisterPredicate(pred string, file string) error {
	plMutex.Lock()
	defer plMutex.Unlock()

	if _, ok := plFacts[pred]; !ok {
		plFacts[pred] = map[string]bool{}
	}
	plFiles[pred] = file

	src, err := os.ReadFile(file)
	if err == nil {
		lines := strings.Split(string(src), "\n")
		prefix := pred + "("
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if strings.HasPrefix(ln, prefix) && strings.HasSuffix(ln, ").") {
				body := ln[len(prefix) : len(ln)-2]
				id := toAtom(body)
				if id != "" {
					plFacts[pred][id] = true
				}
			}
		}
	}
	plRebuildMachine()
	return nil
}

func PLList(pred string) []string {
	plMutex.Lock()
	defer plMutex.Unlock()
	q := pred + "(Id)."
	sols := plMachine.ProveAll(q)
	out := make([]string, 0, len(sols))
	for _, s := range sols {
		out = append(out, fmt.Sprint(s.ByName_("Id")))
	}
	return out
}

func PLCreate(pred, raw string) (string, bool) {
	plMutex.Lock()
	defer plMutex.Unlock()
	id := toAtom(raw)
	if id == "" {
		id = "x"
	}
	if _, ok := plFacts[pred]; !ok {
		plFacts[pred] = map[string]bool{}
	}
	if plFacts[pred][id] {
		return id, false
	}
	plFacts[pred][id] = true
	plRebuildMachine()
	_ = plSave(pred)
	return id, true
}

func PLDelete(pred, raw string) bool {
	plMutex.Lock()
	defer plMutex.Unlock()
	id := toAtom(raw)
	if _, ok := plFacts[pred]; !ok || !plFacts[pred][id] {
		return false
	}
	delete(plFacts[pred], id)
	plRebuildMachine()
	_ = plSave(pred)
	return true
}

func PLUpdate(pred, oldRaw, newRaw string) (string, bool, string) {
	plMutex.Lock()
	defer plMutex.Unlock()
	oldID := toAtom(oldRaw)
	newID := toAtom(newRaw)
	if _, ok := plFacts[pred]; !ok || !plFacts[pred][oldID] {
		return "", false, "not_found"
	}
	if newID == oldID {
		return newID, true, ""
	}
	if plFacts[pred][newID] {
		return "", false, "conflict"
	}
	delete(plFacts[pred], oldID)
	plFacts[pred][newID] = true
	plRebuildMachine()
	_ = plSave(pred)
	return newID, true, ""
}
