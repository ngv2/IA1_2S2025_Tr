package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

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

	// existentes
	http.HandleFunc("/load", withCORS(handleLoad))
	http.HandleFunc("/add", withCORS(handleAddFact))
	http.HandleFunc("/query", withCORS(handleQuery))
	http.HandleFunc("/download", withCORS(handleDownload))

	// === nuevos endpoints para el frontend ===
	http.HandleFunc("/api/symptoms", withCORS(handleListSymptoms))
	http.HandleFunc("/api/medications", withCORS(handleListMedications))
	http.HandleFunc("/api/chronic-conditions", withCORS(handleListChronics))
	http.HandleFunc("/api/diagnosis", withCORS(handleDiagnosis))

	fmt.Println("Servidor iniciado en http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}

// ----------------------------------------------------
// Utilidades
// ----------------------------------------------------

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

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func serverError(w http.ResponseWriter, msg string, err error) {
	fmt.Println(msg, "=>", err)
	http.Error(w, msg, http.StatusInternalServerError)
}

func toAtom(id string) string {
	// convierte strings arbitrarios en átomos prolog válidos
	s := strings.TrimSpace(strings.ToLower(id))
	re := regexp.MustCompile(`[^\p{L}\p{N}_]+`)
	s = re.ReplaceAllString(s, "_")
	if s == "" {
		s = "x"
	}
	return s
}

func prologListAtoms(vals []string) string {
	// construye "[a,b,c]" con átomos
	if len(vals) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(vals))
	for _, v := range vals {
		items = append(items, toAtom(v))
	}
	return "[" + strings.Join(items, ",") + "]"
}

func qAll(q string) ([]map[string]any, error) {
	sols, err := prologVM.Query(q)
	if err != nil {
		return nil, err
	}
	defer sols.Close()
	var out []map[string]any
	for sols.Next() {
		m := make(map[string]any)
		if err := sols.Scan(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := sols.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ----------------------------------------------------
// Handlers existentes (load/add/query/download)
// ----------------------------------------------------

func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(w, "Error al leer el cuerpo de la solicitud", err)
		return
	}

	initProlog()
	code.Write(body)

	if err := prologVM.Exec(code.String()); err != nil {
		serverError(w, "Error al cargar el código Prolog", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Código Prolog cargado exitosamente"))
}

func handleAddFact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(w, "Error al leer el cuerpo de la solicitud", err)
		return
	}

	fact := strings.TrimSpace(string(body))
	if fact == "" || !strings.Contains(fact, "(") {
		badRequest(w, "Hecho o regla inválido")
		return
	}
	pred := fact[:strings.Index(fact, "(")]

	lines := strings.Split(code.String(), "\n")
	var newCode []string
	inserted := false
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, pred+"(") && !inserted {
			newCode = append(newCode, fact, ln)
			inserted = true
		} else {
			newCode = append(newCode, ln)
		}
	}
	if !inserted {
		newCode = append(newCode, fact)
	}
	code.Reset()
	code.WriteString(strings.Join(newCode, "\n"))

	if err := prologVM.Exec(code.String()); err != nil {
		serverError(w, "Error al agregar el hecho", err)
		return
	}
	w.Write([]byte("Hecho agregado correctamente"))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	var q struct{ Query string `json:"query"` }
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		badRequest(w, "Consulta inválida")
		return
	}

	results, err := qAll(q.Query)
	if err != nil {
		serverError(w, "Error en la consulta", err)
		return
	}
	if len(results) == 0 {
		results = append(results, map[string]any{"message": "No se encontraron resultados"})
	} else if len(results) == 1 && len(results[0]) == 0 {
		results[0] = map[string]any{"message": "La consulta es correcta"}
	}
	jsonOK(w, map[string]any{"results": results})
}

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

// ----------------------------------------------------
// NUEVOS: catálogos y diagnóstico para el frontend
// ----------------------------------------------------

type symptomDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleListSymptoms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	rows, err := qAll(`sintoma(Id, Name).`)
	if err != nil {
		serverError(w, "Error listando síntomas", err)
		return
	}
	out := make([]symptomDTO, 0, len(rows))
	for _, m := range rows {
		id, _ := m["Id"].(string)
		name, _ := m["Name"].(string)
		out = append(out, symptomDTO{ID: id, Name: name})
	}
	jsonOK(w, out)
}

type medDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleListMedications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	rows, err := qAll(`medicamento(Id, Name).`)
	if err != nil {
		serverError(w, "Error listando medicamentos", err)
		return
	}
	out := make([]medDTO, 0, len(rows))
	for _, m := range rows {
		id, _ := m["Id"].(string)
		name, _ := m["Name"].(string)
		out = append(out, medDTO{ID: id, Name: name})
	}
	jsonOK(w, out)
}

type chronicDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleListChronics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	rows, err := qAll(`cronica(Id).`)
	if err != nil {
		serverError(w, "Error listando condiciones crónicas", err)
		return
	}
	out := make([]chronicDTO, 0, len(rows))
	for _, m := range rows {
		id, _ := m["Id"].(string)
		// si no hay nombre explícito, usa el id “bonito”
		name := strings.Title(strings.ReplaceAll(id, "_", " "))
		out = append(out, chronicDTO{ID: id, Name: name})
	}
	jsonOK(w, out)
}

// -------- DIAGNOSIS --------

type diagnosisInput struct {
	Symptoms         []struct {
		ID       string `json:"id"`
		Severity string `json:"severity"` // leve | moderado | severo
	} `json:"symptoms"`
	Allergies        []string `json:"allergies"`
	ChronicConditions []string `json:"chronicConditions"`
}

type diagnosisResult struct {
	DiseaseID   string   `json:"diseaseId"`
	DiseaseName string   `json:"diseaseName"`
	Affinity    float64  `json:"affinity"` // 0..1
	Medication  *medDTO  `json:"medication,omitempty"`
	Urgency     string   `json:"urgency"`
	Conflicts   []string `json:"conflicts,omitempty"`
}

type diagnosisResponse struct {
	GeneratedAt string            `json:"generatedAt"`
	RulesGlobal string            `json:"rulesGlobal"`
	Results     []diagnosisResult `json:"results"`
}

func handleDiagnosis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()

	var in diagnosisInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "Cuerpo inválido")
		return
	}
	if len(in.Symptoms) == 0 {
		badRequest(w, "Debes enviar al menos un síntoma")
		return
	}

	// 1) listar todas las enfermedades conocidas
	rows, err := qAll(`enfermedad(E, Name).`)
	if err != nil {
		serverError(w, "Error consultando enfermedades", err)
		return
	}
	if len(rows) == 0 {
		jsonOK(w, diagnosisResponse{GeneratedAt: time.Now().UTC().Format(time.RFC3339), Results: []diagnosisResult{}})
		return
	}

	// map severidad -> factor
	severityFactor := map[string]float64{
		"leve":     0.8,
		"moderado": 1.0,
		"severo":   1.2,
	}

	// preparar listas para prolog (urgencia y medicamentos)
	severitiesAtoms := make([]string, 0, len(in.Symptoms))
	for _, s := range in.Symptoms {
		severitiesAtoms = append(severitiesAtoms, toAtom(s.Severity))
	}
	allergiesAtoms := make([]string, 0, len(in.Allergies))
	for _, a := range in.Allergies {
		allergiesAtoms = append(allergiesAtoms, toAtom(a))
	}
	chronicsAtoms := make([]string, 0, len(in.ChronicConditions))
	for _, c := range in.ChronicConditions {
		chronicsAtoms = append(chronicsAtoms, toAtom(c))
	}

	// 2) calcular afinidad y medicación por enfermedad
	results := make([]diagnosisResult, 0, len(rows))

	for _, m := range rows {
		eID, _ := m["E"].(string)
		eName, _ := m["Name"].(string)

		total := 0.0
		for _, s := range in.Symptoms {
			// consulta peso: enfermedad_sintoma(E, S, Peso).
			q := fmt.Sprintf(`enfermedad_sintoma(%s, %s, P).`, toAtom(eID), toAtom(s.ID))
			ps, err := qAll(q)
			if err != nil {
				serverError(w, "Error consultando pesos", err)
				return
			}
			if len(ps) > 0 {
				// ichiban entrega números como float64
				p, _ := ps[0]["P"].(float64)
				f := severityFactor[strings.ToLower(s.Severity)]
				if f == 0 {
					f = 1.0
				}
				total += p * f
			}
		}
		if total > 1.0 {
			total = 1.0
		}

		// urgencia desde Prolog: urgencia([leve,moderado,...], U).
		urgQ := fmt.Sprintf(`urgencia(%s, U).`, prologListAtoms(severitiesAtoms))
		urgRows, _ := qAll(urgQ)
		urgency := "Observación recomendada"
		if len(urgRows) > 0 {
			if u, ok := urgRows[0]["U"].(string); ok && u != "" {
				urgency = u
			}
		}

		// sugerir medicamento seguro
		medQ := fmt.Sprintf(`sugerir_medicamento(%s, %s, %s, M).`,
			toAtom(eID),
			prologListAtoms(allergiesAtoms),
			prologListAtoms(chronicsAtoms),
		)
		medRows, _ := qAll(medQ)

		var med *medDTO
		var conflicts []string

		if len(medRows) > 0 {
			if mid, ok := medRows[0]["M"].(string); ok {
				// obtener nombre
				nQ := fmt.Sprintf(`medicamento(%s, N).`, toAtom(mid))
				nRows, _ := qAll(nQ)
				name := strings.Title(strings.ReplaceAll(mid, "_", " "))
				if len(nRows) > 0 {
					if n, ok := nRows[0]["N"].(string); ok && n != "" {
						name = n
					}
				}
				med = &medDTO{ID: mid, Name: name}
			}
		} else {
			// no hay opción segura -> explicar conflictos de la primera opción candidata
			// encuentra tratamientos y marca conflictos por alergias o contraindicaciones
			cQ := fmt.Sprintf(
				`trata(%s, M), ( member(M, %s) ; (contraindicacion(M, C), member(C, %s)) ).`,
				toAtom(eID),
				prologListAtoms(allergiesAtoms),
				prologListAtoms(chronicsAtoms),
			)
			cRows, _ := qAll(cQ)
			for _, r := range cRows {
				if mID, ok := r["M"].(string); ok {
					conflicts = append(conflicts, "Conflicto con: "+mID)
				}
			}
		}

		results = append(results, diagnosisResult{
			DiseaseID:   eID,
			DiseaseName: eName,
			Affinity:    round2(total),
			Medication:  med,
			Urgency:     urgency,
			Conflicts:   conflicts,
		})
	}

	// ordenar por afinidad (desc) en el frontend; aquí devolvemos tal cual
	resp := diagnosisResponse{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		RulesGlobal: "", // si quieres, puedes construir aquí un resumen de reglas activadas
		Results:     results,
	}
	jsonOK(w, resp)
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100.0
}
