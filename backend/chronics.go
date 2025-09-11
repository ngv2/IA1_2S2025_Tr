package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const predChronics = "cronica"
const fileChronics = "prolog.pl"

type chronicDTO struct {
	ID string `json:"id"`
}

func InitChronics() error {
	return PLRegisterPredicate(predChronics, fileChronics)
}

func ListChronics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ids := PLList(predChronics)
	out := make([]chronicDTO, 0, len(ids))
	for _, id := range ids {
		out = append(out, chronicDTO{ID: id})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func CreateChronic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body chronicDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"...\"}"})
		return
	}
	id, ok := PLCreate(predChronics, body.ID)
	if !ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "La crónica ya existe"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chronicDTO{ID: id})
}

func DeleteChronic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/chronics/{id}"})
		return
	}
	id := parts[2]
	if !PLDelete(predChronics, id) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "No existe la crónica"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func UpdateChronic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/chronics/{id}"})
		return
	}
	oldID := parts[2]
	var body chronicDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"nuevo_id\"}"})
		return
	}
	newID, ok, why := PLUpdate(predChronics, oldID, body.ID)
	if !ok {
		switch why {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(apiError{Error: "No existe la crónica a actualizar"})
		case "conflict":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "Ya existe una crónica con ese id"})
		default:
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chronicDTO{ID: newID})
}
