package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Uso:")
		fmt.Println("  rpa_prolog_loader.exe <ruta_prolog_pl> [ruta_texto_para_insertar]")
		os.Exit(2)
	}

	plPath := mustAbs(os.Args[1])
	var payload string
	if len(os.Args) == 3 {
		payload = string(mustReadFile(os.Args[2]))
	}

	openNotepadWithFile(plPath)
	focusNotepad(20, 200)
	time.Sleep(800 * time.Millisecond)

	if payload != "" {
		robotgo.KeyTap("end", "ctrl")
		time.Sleep(120 * time.Millisecond)
		robotgo.KeyTap("enter")
		time.Sleep(100 * time.Millisecond)
		robotgo.WriteAll(payload)
		time.Sleep(80 * time.Millisecond)
		robotgo.KeyTap("v", "ctrl")
		time.Sleep(200 * time.Millisecond)
		robotgo.KeyTap("s", "ctrl")
		time.Sleep(400 * time.Millisecond)
	}

	fmt.Println("Listo.")
}

func openNotepadWithFile(path string) {
	_ = exec.Command("cmd", "/C", "start", "", "notepad.exe", path).Start()
}

func focusNotepad(retries int, delayMs int) {
	for i := 0; i < retries; i++ {
		_ = robotgo.ActiveName("Notepad")
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}

func mustAbs(p string) string {
	ap, err := filepath.Abs(p)
	if err != nil {
		fmt.Println("Error resolviendo ruta:", err)
		os.Exit(1)
	}
	return ap
}

func mustReadFile(p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		fmt.Println("Error leyendo archivo:", p, "=>", err)
		os.Exit(1)
	}
	return b
}
