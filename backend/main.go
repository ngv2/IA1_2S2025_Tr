package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ichiban/prolog"
)

var (
	prologVM *prolog.Interpreter
	code     bytes.Buffer
	mutex    sync.Mutex
)

func initProlog() {
	prologVM = prolog.New(os.Stdin, os.Stdout)
	code.Reset()
}

func main() {
	initProlog()

	http.HandleFunc("/load", withCORS(handleLoad))
	http.HandleFunc("/add", withCORS(handleAddFact))
	http.HandleFunc("/query", withCORS(handleQuery))
	http.HandleFunc("/download", withCORS(handleDownload))

	fmt.Println("Servidor iniciado en http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}

// Midleware para CORS
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// POST /load - carga el código prolog desde un archivo de texto
func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error al leer el cuerpo de la solicitud:", err)
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	initProlog()
	code.Write(body)

	err = prologVM.Exec(code.String())
	if err != nil {
		fmt.Println("Error al cargar el código Prolog:", err)
		http.Error(w, "Error al cargar el código Prolog", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Código Prolog cargado exitosamente"))
}

// POST /add - agrega un hecho al código Prolog
func handleAddFact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error al leer el cuerpo de la solicitud:", err)
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	fact := string(body)
	fact = strings.TrimSpace(fact)
	if fact == "" {
		http.Error(w, "Hecho o regla inválido", http.StatusBadRequest)
		return
	}

	openParen := strings.Index(fact, "(")
	if openParen == -1 {
		http.Error(w, "Hecho o regla inválido", http.StatusBadRequest)
		return
	}
	pred := fact[:openParen]

	lines := strings.Split(code.String(), "\n")
	var newCodeLines []string

	inserted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, pred+"(") && !inserted {
			newCodeLines = append(newCodeLines, fact)
			newCodeLines = append(newCodeLines, line)
			inserted = true
		} else {
			newCodeLines = append(newCodeLines, line)
		}
	}
	if !inserted {
		newCodeLines = append(newCodeLines, fact)
	}

	code.Reset()
	code.WriteString(strings.Join(newCodeLines, "\n"))

	err = prologVM.Exec(code.String())
	if err != nil {
		fmt.Println("Error al agregar el hecho:", err)
		http.Error(w, fmt.Sprintf("Error al agregar el hecho: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hecho agregado correctamente"))
}

// POST /query - ejecuta una consulta Prolog
func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	var query struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		fmt.Println("Error al leer el cuerpo de la solicitud:", err)
		http.Error(w, "Consulta inválida", http.StatusBadRequest)
		return
	}

	solutions, err := prologVM.Query(query.Query)
	if err != nil {
		fmt.Println("Error en la consulta Prolog:", err)
		http.Error(w, fmt.Sprintf("Error en la consulta %v", err), http.StatusInternalServerError)
		return
	}
	defer solutions.Close()

	var results []map[string]any

	for solutions.Next() {
		m := make(map[string]any)
		if err := solutions.Scan(&m); err != nil {
			fmt.Println("Error al escanear resultados:", err)
			http.Error(w, fmt.Sprintf("Error al escanear resultados: %v", err), http.StatusInternalServerError)
			return
		}
		results = append(results, m)
	}

	if err := solutions.Err(); err != nil {
		fmt.Println("Error en la consulta Prolog:", err)
		http.Error(w, fmt.Sprintf("Error en la consulta %v", err), http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		results = append(results, map[string]any{"message": "No se encontraron resultados"})
	} else if len(results) == 1 && len(results[0]) == 0 {
		results[0] = map[string]any{"message": "La consulta es correcta"}
	}

	resp := map[string]interface{}{
		"results": results,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /download - descarga el código Prolog actual
func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	w.Header().Set("Content-Disposition", "attachment; filename=code.pl")
	w.Header().Set("Content-Type", "text/plain")
	w.Write(code.Bytes())
}
