package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const predSymptoms = "sintoma"
const fileSymptoms = "prolog.pl"

type symptomDTO struct {
	ID string `json:"id"`
}

type apiError struct {
	Error string `json:"error"`
}

func InitSymptoms() error {
	return PLRegisterPredicate(predSymptoms, fileSymptoms)
}

func listSymptoms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ids := PLList(predSymptoms)
	out := make([]symptomDTO, 0, len(ids))
	for _, id := range ids {
		out = append(out, symptomDTO{ID: id})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func createSymptom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body symptomDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"...\"}"})
		return
	}
	id, ok := PLCreate(predSymptoms, body.ID)
	if !ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "El síntoma ya existe"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(symptomDTO{ID: id})
}

func deleteSymptom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/symptoms/{id}"})
		return
	}
	id := parts[2]
	if !PLDelete(predSymptoms, id) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "No existe el síntoma"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func updateSymptom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/symptoms/{id}"})
		return
	}
	oldID := parts[2]
	var body symptomDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"nuevo_id\"}"})
		return
	}
	newID, ok, why := PLUpdate(predSymptoms, oldID, body.ID)
	if !ok {
		switch why {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(apiError{Error: "No existe el síntoma a actualizar"})
		case "conflict":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "Ya existe un síntoma con ese id"})
		default:
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symptomDTO{ID: newID})
}
