package main

import (
	"fmt"
	"net/http"
)

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	if err := InitSymptoms(); err != nil { panic(err) }
	if err := InitDiseases(); err != nil { panic(err) }
	if err := InitMedications(); err != nil { panic(err) }
	if err := InitChronics(); err != nil { panic(err) }
	if err := InitAllergies(); err != nil { panic(err) }

	http.HandleFunc("/api/symptoms", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:  listSymptoms(w,r)
		case http.MethodPost: createSymptom(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/symptoms/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: deleteSymptom(w,r)
		case http.MethodPut, http.MethodPatch: updateSymptom(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/diseases", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:  ListDiseases(w,r)
		case http.MethodPost: CreateDisease(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/diseases/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: DeleteDisease(w,r)
		case http.MethodPut, http.MethodPatch: UpdateDisease(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/medications", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:  ListMedications(w,r)
		case http.MethodPost: CreateMedication(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/medications/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: DeleteMedication(w,r)
		case http.MethodPut, http.MethodPatch: UpdateMedication(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/chronics", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:  ListChronics(w,r)
		case http.MethodPost: CreateChronic(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/chronics/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: DeleteChronic(w,r)
		case http.MethodPut, http.MethodPatch: UpdateChronic(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/allergies", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:  ListAllergies(w,r)
		case http.MethodPost: CreateAllergy(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/allergies/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: DeleteAllergy(w,r)
		case http.MethodPut, http.MethodPatch: UpdateAllergy(w,r)
		default: http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/diagnosis", withCORS(handleDiagnosis))


	fmt.Println("Servidor en http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}
