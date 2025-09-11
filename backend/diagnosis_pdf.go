package main

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

func handleDiagnosisPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var in DiagnosisIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || len(in.Symptoms) == 0 {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	out := runDiagnosisReport(in)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Informe de Diagnostico")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, "Fecha: "+time.Now().Format("2006-01-02 15:04"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, "Resumen de entrada")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, formatInputs(out.Inputs), "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, "Resultados")
	pdf.Ln(8)

	header := []string{"Enfermedad", "Afinidad", "Urgencia", "Medicamento"}
	colW := []float64{60, 60, 40, 25}
	pdf.SetFont("Arial", "B", 11)
	for i, h := range header {
		pdf.CellFormat(colW[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	for _, rls := range out.Results {
		pdf.CellFormat(colW[0], 8, rls.DiseaseName, "1", 0, "L", false, 0, "")
		px, py := pdf.GetXY()
		pdf.CellFormat(colW[1], 8, "", "1", 0, "", false, 0, "")
		pdf.SetXY(px, py)
		drawAffinityBar(pdf, px+1.5, py+1.5, colW[1]-3, 5, rls.Affinity)
		pdf.SetXY(px, py)
		pdf.CellFormat(colW[1], 8, strconv.Itoa(int(math.Round(rls.Affinity*100)))+"%", "", 0, "R", false, 0, "")
		pdf.CellFormat(colW[2], 8, rls.Urgency, "1", 0, "C", false, 0, "")
		med := "-"
		if rls.Medication != nil {
			med = rls.Medication.Name
		}
		pdf.CellFormat(colW[3], 8, med, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)

		if len(rls.Conflicts) > 0 {
			pdf.SetFont("Arial", "I", 9)
			pdf.SetTextColor(200, 0, 0)
			pdf.MultiCell(0, 5, "Advertencias: "+strings.Join(rls.Conflicts, "; "), "LRB", "L", false)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 10)
		}

		pdf.SetFont("Arial", "", 9)
		if len(rls.Contributions) > 0 {
			var sb strings.Builder
			sb.WriteString("Contribuciones: ")
			for i, c := range rls.Contributions {
				if i > 0 {
					sb.WriteString(" | ")
				}
				sb.WriteString(c.SymptomID + "(" + c.Severity + "): " + trimFloat(c.Contribution))
			}
			pdf.MultiCell(0, 5, sb.String(), "LRB", "L", false)
		}

		if len(rls.RulesActivated) > 0 {
			var rb strings.Builder
			rb.WriteString("Reglas: ")
			for i, rr := range rls.RulesActivated {
				if i > 0 {
					rb.WriteString(" | ")
				}
				rb.WriteString(rr.Rule + "[" + rr.Details + "]")
			}
			pdf.MultiCell(0, 5, rb.String(), "LRB", "L", false)
		}
		pdf.Ln(2)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, "Notas")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Este informe es informativo y no sustituye una evaluación médica profesional.", "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, "error generando PDF", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=diagnostico.pdf")
	w.Write(buf.Bytes())
}

func runDiagnosisReport(in DiagnosisIn) DiagnosisOut {
	severityFactor := map[string]float64{"leve": 0.8, "moderado": 1.0, "severo": 1.2}

	pairsDiseases := List2(predDiseases)
	triplesES := List3(predDisSym)
	pairsTrata := List2("trata")
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

	urg, urgDetail := computeUrgencyPDF(in.Symptoms)

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

		var mChosen *DxMedication
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
		rules = append(rules, DxRule{Rule: "urgencia/2", Details: urgDetail})

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
	return DiagnosisOut{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Inputs:      in,
		Results:     results,
	}
}

func drawAffinityBar(pdf *gofpdf.Fpdf, x, y, w float64, h float64, affinity float64) {
	bw := w * affinity
	pdf.SetFillColor(33, 150, 243)
	pdf.Rect(x, y, bw, h, "F")
}

func computeUrgencyPDF(syms []DxSymptom) (string, string) {
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

func trimFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func formatInputs(in DiagnosisIn) string {
	var a, c, s []string
	for _, x := range in.Allergies { a = append(a, toAtom(x)) }
	for _, x := range in.Chronics  { c = append(c, toAtom(x)) }
	for _, x := range in.Symptoms  { s = append(s, toAtom(x.ID)+"("+strings.ToLower(x.Severity)+")") }
	return "Sintomas: [" + strings.Join(s, ", ") + "]\nAlergias: [" + strings.Join(a, ", ") + "]\nCronicas: [" + strings.Join(c, ", ") + "]"
}
