package main

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const predTrata = "trata"

type DxSymptom struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
}

type DiagnosisIn struct {
	Symptoms []DxSymptom `json:"symptoms"`
	Allergies []string   `json:"allergies"`
	Chronics  []string   `json:"chronics"`
}

type DxContribution struct {
	SymptomID    string  `json:"symptomId"`
	Severity     string  `json:"severity"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
}

type DxRule struct {
	Rule    string `json:"rule"`
	Details string `json:"details"`
}

type DxMedication struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DxResult struct {
	DiseaseID      string           `json:"diseaseId"`
	DiseaseName    string           `json:"diseaseName"`
	Affinity       float64          `json:"affinity"`
	AffinityPct    float64          `json:"affinityPercent"`
	Urgency        string           `json:"urgency"`
	Medication     *DxMedication    `json:"medication,omitempty"`
	Conflicts      []string         `json:"conflicts,omitempty"`
	Contributions  []DxContribution `json:"contributions"`
	RulesActivated []DxRule         `json:"rulesActivated"`
}

type DiagnosisOut struct {
	GeneratedAt string      `json:"generatedAt"`
	Inputs      DiagnosisIn `json:"inputs"`
	Results     []DxResult  `json:"results"`
}

func InitDiagnosis() error {
	if err := Register2(predTrata, fileDiseases); err != nil {
		return err
	}
	return nil
}

func handleDiagnosis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(apiError{Error: "método no permitido"})
		return
	}
	var in DiagnosisIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "JSON inválido"})
		return
	}
	if len(in.Symptoms) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "debes enviar al menos un síntoma"})
		return
	}

	severityFactor := map[string]float64{"leve": 0.8, "moderado": 1.0, "severo": 1.2}

	pairsDiseases := List2(predDiseases)
	triplesES := List3(predDisSym)
	pairsTrata := List2(predTrata)
	pairsMeds := List2(predMeds)
	pairsContra := List2(predContra)

	weight := map[string]map[string]float64{}
	for _, t := range triplesES {
		e, s := t[0], t[1]
		wf, _ := strconv.ParseFloat(t[2], 64)
		if weight[e] == nil {
			weight[e] = map[string]float64{}
		}
		weight[e][s] = wf
	}

	medName := map[string]string{}
	for _, m := range pairsMeds {
		medName[m[0]] = m[1]
	}

	trata := map[string][]string{}
	for _, p := range pairsTrata {
		trata[p[0]] = append(trata[p[0]], p[1])
	}

	contra := map[[2]string]bool{}
	for _, c := range pairsContra {
		contra[[2]string{c[0], c[1]}] = true
	}

	allergies := map[string]bool{}
	for _, a := range in.Allergies {
		allergies[toAtom(a)] = true
	}
	chronics := map[string]bool{}
	for _, c := range in.Chronics {
		chronics[toAtom(c)] = true
	}

	urg, urgRuleDetail := computeUrgency(in.Symptoms)

	results := make([]DxResult, 0, len(pairsDiseases))
	for _, d := range pairsDiseases {
		eID, eName := d[0], d[1]
		total := 0.0
		var contribs []DxContribution
		var rules []DxRule

		for _, s := range in.Symptoms {
			sid := toAtom(s.ID)
			w := 0.0
			if weight[eID] != nil {
				w = weight[eID][sid]
			}
			f := severityFactor[strings.ToLower(s.Severity)]
			if f == 0 {
				f = 1.0
			}
			c := w * f
			if c > 0 {
				contribs = append(contribs, DxContribution{
					SymptomID: sid, Severity: strings.ToLower(s.Severity), Weight: round2dx(w), Contribution: round2dx(c),
				})
				rules = append(rules, DxRule{Rule: "enfermedad_sintoma/3", Details: eID + "," + sid + "," + strconv.FormatFloat(w, 'g', -1, 64)})
			}
			total += c
		}
		if total > 1.0 {
			total = 1.0
		}

		mChosen := (*DxMedication)(nil)
		var conflicts []string
		for _, m := range trata[eID] {
			if allergies[m] {
				conflicts = append(conflicts, "alergia:"+m)
				continue
			}
			conflict := false
			for ch := range chronics {
				if contra[[2]string{m, ch}] {
					conflicts = append(conflicts, "contra:"+m+"-"+ch)
					conflict = true
					break
				}
			}
			if conflict {
				continue
			}
			name := medName[m]
			if name == "" {
				name = m
			}
			mChosen = &DxMedication{ID: m, Name: name}
			rules = append(rules, DxRule{Rule: "trata/2", Details: eID + "," + m})
			break
		}
		if mChosen == nil && len(trata[eID]) > 0 && len(conflicts) > 0 {
			rules = append(rules, DxRule{Rule: "exclusion_tratamiento", Details: strings.Join(conflicts, ";")})
		}
		rules = append(rules, DxRule{Rule: "urgencia/2", Details: urgRuleDetail})

		results = append(results, DxResult{
			DiseaseID:      eID,
			DiseaseName:    eName,
			Affinity:       round2dx(total),
			AffinityPct:    round2dx(total * 100),
			Urgency:        urg,
			Medication:     mChosen,
			Conflicts:      conflicts,
			Contributions:  contribs,
			RulesActivated: rules,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Affinity > results[j].Affinity })

	out := DiagnosisOut{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Inputs:      in,
		Results:     results,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func computeUrgency(syms []DxSymptom) (string, string) {
	hasSevere := false
	hasModerate := false
	var sevs []string
	for _, s := range syms {
		ss := strings.ToLower(strings.TrimSpace(s.Severity))
		sevs = append(sevs, toAtom(ss))
		if ss == "severo" {
			hasSevere = true
		} else if ss == "moderado" {
			hasModerate = true
		}
	}
	if hasSevere {
		return "consulta_medica_inmediata_sugerida", strings.Join(sevs, ",") + "->consulta_medica_inmediata_sugerida"
	}
	if hasModerate {
		return "observacion_recomendada", strings.Join(sevs, ",") + "->observacion_recomendada"
	}
	return "posible_automanejo", strings.Join(sevs, ",") + "->posible_automanejo"
}

func listAtoms(vals []string) string {
	if len(vals) == 0 {
		return "[]"
	}
	items := make([]string, len(vals))
	for i, v := range vals {
		items[i] = toAtom(v)
	}
	return "[" + strings.Join(items, ",") + "]"
}

func round2dx(v float64) float64 {
	return math.Round(v*100) / 100
}
