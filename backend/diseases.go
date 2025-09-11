package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	predDiseases = "enfermedad"
	predDisSym   = "enfermedad_sintoma"
	fileDiseases = "prolog.pl"
	fileDisSym   = "prolog.pl"
)

type DiseaseSym struct {
	ID     string  `json:"id"`
	Weight float64 `json:"weight"`
}

type DiseaseIn struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Symptoms []DiseaseSym `json:"symptoms,omitempty"`
}

type DiseaseOut struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Symptoms []DiseaseSym `json:"symptoms"`
}

func InitDiseases() error {
	if err := Register2(predDiseases, fileDiseases); err != nil {
		return err
	}
	if err := Register3(predDisSym, fileDisSym); err != nil {
		return err
	}
	return nil
}

func ListDiseases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	pairs := List2(predDiseases)
	tris := List3(predDisSym)
	out := make([]DiseaseOut, 0, len(pairs))
	for _, p := range pairs {
		enfID := p[0]
		enfName := p[1]
		var syms []DiseaseSym
		for _, t := range tris {
			if t[0] == enfID {
				wf, _ := strconv.ParseFloat(t[2], 64)
				syms = append(syms, DiseaseSym{ID: t[1], Weight: wf})
			}
		}
		out = append(out, DiseaseOut{ID: enfID, Name: enfName, Symptoms: syms})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func CreateDisease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	var in DiseaseIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido"})
		return
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "id y name son obligatorios"})
		return
	}
	if _, ok := Create2(predDiseases, in.ID, in.Name); !ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "la enfermedad ya existe"})
		return
	}
	for _, s := range in.Symptoms {
		ws := strconv.FormatFloat(s.Weight, 'g', -1, 64)
		_, _ = Create3(predDisSym, in.ID, s.ID, ws)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(DiseaseOut{ID: toAtom(in.ID), Name: toAtom(in.Name), Symptoms: in.Symptoms})
}

func UpdateDisease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "ruta: /api/diseases/{oldId}"})
		return
	}
	oldID := parts[2]
	var in DiseaseIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido"})
		return
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "id y name son obligatorios"})
		return
	}
	oldName := getDiseaseName(oldID)
	if _, ok, why := Update2(predDiseases, oldID, oldName, in.ID, in.Name); !ok {
		switch why {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(apiError{Error: "no existe la enfermedad a actualizar"})
		case "conflict":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "ya existe una enfermedad con ese id/nombre"})
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(apiError{Error: "error en actualización"})
		}
		return
	}
	if toAtom(oldID) != toAtom(in.ID) {
		renameDiseaseInTriples(oldID, in.ID)
	}
	if in.Symptoms != nil {
		replaceDiseaseSymptoms(in.ID, in.Symptoms)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DiseaseOut{ID: toAtom(in.ID), Name: toAtom(in.Name), Symptoms: readDiseaseSymptoms(in.ID)})
}

func DeleteDisease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "ruta: /api/diseases/{id}"})
		return
	}
	id := parts[2]
	if !Delete2(predDiseases, id, getDiseaseName(id)) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "no existe la enfermedad"})
		return
	}
	deleteAllDiseaseTriples(id)
	w.WriteHeader(http.StatusNoContent)
}

func getDiseaseName(id string) string {
	id = toAtom(id)
	for _, p := range List2(predDiseases) {
		if p[0] == id {
			return p[1]
		}
	}
	return ""
}

func readDiseaseSymptoms(diseaseID string) []DiseaseSym {
	diseaseID = toAtom(diseaseID)
	var syms []DiseaseSym
	for _, t := range List3(predDisSym) {
		if t[0] == diseaseID {
			wf, _ := strconv.ParseFloat(t[2], 64)
			syms = append(syms, DiseaseSym{ID: t[1], Weight: wf})
		}
	}
	return syms
}

func deleteAllDiseaseTriples(diseaseID string) {
	diseaseID = toAtom(diseaseID)
	for _, t := range List3(predDisSym) {
		if t[0] == diseaseID {
			_ = Delete3(predDisSym, t[0], t[1], t[2])
		}
	}
}

func replaceDiseaseSymptoms(diseaseID string, list []DiseaseSym) {
	diseaseID = toAtom(diseaseID)
	deleteAllDiseaseTriples(diseaseID)
	for _, s := range list {
		ws := strconv.FormatFloat(s.Weight, 'g', -1, 64)
		_, _ = Create3(predDisSym, diseaseID, s.ID, ws)
	}
}

func renameDiseaseInTriples(oldID, newID string) {
	oldID = toAtom(oldID)
	newID = toAtom(newID)
	for _, t := range List3(predDisSym) {
		if t[0] == oldID {
			_, ok, _ := Update3(predDisSym, t[0], t[1], t[2], newID, t[1], t[2])
			if !ok {
				_ = Delete3(predDisSym, t[0], t[1], t[2])
			}
		}
	}
}
