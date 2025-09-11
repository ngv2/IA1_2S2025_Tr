package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const predAllergies = "alergia"
const fileAllergies = "prolog.pl"

type allergyDTO struct {
	ID string `json:"id"`
}

func InitAllergies() error {
	return PLRegisterPredicate(predAllergies, fileAllergies)
}

func ListAllergies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ids := PLList(predAllergies)
	out := make([]allergyDTO, 0, len(ids))
	for _, id := range ids {
		out = append(out, allergyDTO{ID: id})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func CreateAllergy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body allergyDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"...\"}"})
		return
	}
	id, ok := PLCreate(predAllergies, body.ID)
	if !ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "La alergia ya existe"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(allergyDTO{ID: id})
}

func DeleteAllergy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/allergies/{id}"})
		return
	}
	id := parts[2]
	if !PLDelete(predAllergies, id) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "No existe la alergia"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func UpdateAllergy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "Ruta inválida. Usa /api/allergies/{id}"})
		return
	}
	oldID := parts[2]
	var body allergyDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido. Envía {\"id\":\"nuevo_id\"}"})
		return
	}
	newID, ok, why := PLUpdate(predAllergies, oldID, body.ID)
	if !ok {
		switch why {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(apiError{Error: "No existe la alergia a actualizar"})
		case "conflict":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "Ya existe una alergia con ese id"})
		default:
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allergyDTO{ID: newID})
}
