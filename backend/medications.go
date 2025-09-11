package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	predMeds   = "medicamento"
	predContra = "contraindicacion"
	fileMeds   = "prolog.pl"
	fileContra = "prolog.pl"
)

type MedicationIn struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Contraindications []string `json:"contraindications,omitempty"`
}

type MedicationOut struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Contraindications []string `json:"contraindications"`
}

func InitMedications() error {
	if err := Register2(predMeds, fileMeds); err != nil {
		return err
	}
	if err := Register2(predContra, fileContra); err != nil {
		return err
	}
	return nil
}

func ListMedications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	meds := List2(predMeds)
	cons := List2(predContra)
	out := make([]MedicationOut, 0, len(meds))
	for _, m := range meds {
		id := m[0]
		name := m[1]
		var cs []string
		for _, c := range cons {
			if c[0] == id {
				cs = append(cs, c[1])
			}
		}
		out = append(out, MedicationOut{ID: id, Name: name, Contraindications: cs})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func CreateMedication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	var in MedicationIn
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
	if _, ok := Create2(predMeds, in.ID, in.Name); !ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "el medicamento ya existe"})
		return
	}
	for _, ch := range in.Contraindications {
		_, _ = Create2(predContra, in.ID, ch)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(MedicationOut{ID: toAtom(in.ID), Name: toAtom(in.Name), Contraindications: in.Contraindications})
}

func UpdateMedication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "ruta: /api/medications/{oldId}"})
		return
	}
	oldID := parts[2]
	var in MedicationIn
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
	oldName := getMedicationName(oldID)
	if _, ok, why := Update2(predMeds, oldID, oldName, in.ID, in.Name); !ok {
		switch why {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(apiError{Error: "no existe el medicamento a actualizar"})
		case "conflict":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "ya existe un medicamento con ese id/nombre"})
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(apiError{Error: "error en actualización"})
		}
		return
	}
	if toAtom(oldID) != toAtom(in.ID) {
		renameContraForMedication(oldID, in.ID)
	}
	if in.Contraindications != nil {
		replaceMedicationContra(in.ID, in.Contraindications)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MedicationOut{ID: toAtom(in.ID), Name: toAtom(in.Name), Contraindications: readMedicationContra(in.ID)})
}

func DeleteMedication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "ruta: /api/medications/{id}"})
		return
	}
	id := parts[2]
	if !Delete2(predMeds, id, getMedicationName(id)) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "no existe el medicamento"})
		return
	}
	deleteAllMedicationContra(id)
	w.WriteHeader(http.StatusNoContent)
}

func getMedicationName(id string) string {
	id = toAtom(id)
	for _, p := range List2(predMeds) {
		if p[0] == id {
			return p[1]
		}
	}
	return ""
}

func readMedicationContra(medID string) []string {
	medID = toAtom(medID)
	var out []string
	for _, c := range List2(predContra) {
		if c[0] == medID {
			out = append(out, c[1])
		}
	}
	return out
}

func deleteAllMedicationContra(medID string) {
	medID = toAtom(medID)
	for _, c := range List2(predContra) {
		if c[0] == medID {
			_ = Delete2(predContra, c[0], c[1])
		}
	}
}

func replaceMedicationContra(medID string, list []string) {
	medID = toAtom(medID)
	deleteAllMedicationContra(medID)
	for _, ch := range list {
		_, _ = Create2(predContra, medID, ch)
	}
}

func renameContraForMedication(oldID, newID string) {
	oldID = toAtom(oldID)
	newID = toAtom(newID)
	for _, c := range List2(predContra) {
		if c[0] == oldID {
			_, ok, _ := Update2(predContra, c[0], c[1], newID, c[1])
			if !ok {
				_ = Delete2(predContra, c[0], c[1])
			}
		}
	}
}
