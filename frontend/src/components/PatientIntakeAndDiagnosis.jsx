import { useEffect, useMemo, useState } from "react";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";
import jsPDF from "jspdf";
import html2canvas from "html2canvas";

/**
 * Vista: Módulo de Pacientes – Ingreso de datos y análisis basado en reglas
 *
 * Cambios clave (UX y robustez):
 * - ✦ Se ocultan errores técnicos y se muestran mensajes genéricos y amables al usuario.
 * - ✦ Detección de respuestas no‑JSON del backend (evita "Unexpected token '<'").
 * - ✦ Skeletons/estados de carga y vacíos mejorados.
 * - ✦ Estilos propios (sin Tailwind) inyectados desde el componente para no depender de setup.
 * - ✦ Layout limpio y consistente.
 */

// === CSS minimalista, inyectado para no depender de frameworks ===
const STYLE_TAG_ID = "patient-intake-styles";
const BASE_CSS = `
:root{--bg:#f6f8fb;--txt:#0f172a;--muted:#64748b;--card:#ffffff;--line:#e2e8f0;--brand:#2563eb;--brand-2:#0ea5e9;--good:#059669;--warn:#ca8a04;--bad:#dc2626}
*{box-sizing:border-box}
body{color:var(--txt)}
.pi-wrap{min-height:100vh;background:var(--bg)}
.pi-container{max-width:1120px;margin:0 auto;padding:24px}
.pi-header h1{font-size:28px;margin:0 0 6px}
.pi-header p{color:var(--muted);margin:0}
.pi-banner{margin:16px 0;padding:10px 12px;border:1px solid #fecaca;background:#fee2e2;color:#7f1d1d;border-radius:10px}
.pi-grid{display:grid;grid-template-columns:1fr;gap:16px;margin-bottom:24px}
@media(min-width:900px){.pi-grid{grid-template-columns:2fr 1fr}}
.pi-card{background:var(--card);border:1px solid var(--line);border-radius:16px;box-shadow:0 1px 2px rgba(0,0,0,.05);padding:16px}
.pi-title{font-size:18px;margin:0 0 12px}
.pi-muted{color:var(--muted);font-size:13px}
.pi-symptoms{display:grid;grid-template-columns:1fr;gap:10px}
@media(min-width:600px){.pi-symptoms{grid-template-columns:1fr 1fr}}
@media(min-width:1000px){.pi-symptoms{grid-template-columns:1fr 1fr 1fr}}
.pi-chip{border:1px solid var(--line);border-radius:12px;padding:10px}
.pi-chip.active{border-color:#93c5fd;background:#eff6ff}
.pi-chip label{display:flex;gap:8px;align-items:flex-start;cursor:pointer}
.pi-chip .desc{font-size:12px;color:var(--muted)}
.pi-row{display:flex;align-items:center;gap:8px}
.pi-select{font-size:13px;padding:6px 8px;border-radius:8px;border:1px solid var(--line)}
.pi-list{max-height:220px;overflow:auto;border:1px solid var(--line);border-radius:12px;padding:8px}
.pi-actions{display:flex;gap:10px;align-items:center;margin:8px 0 24px}
.pi-btn{border:1px solid var(--line);background:#fff;padding:8px 14px;border-radius:12px;cursor:pointer}
.pi-btn.primary{background:var(--brand);color:#fff;border-color:transparent}
.pi-btn.primary:disabled{opacity:.6;cursor:not-allowed}
.pi-btn.secondary{background:#fff}
.pi-btn.success{background:var(--good);color:#fff;border-color:transparent;margin-left:auto}
.pi-table{width:100%;border-collapse:collapse;font-size:14px}
.pi-table th,.pi-table td{padding:10px;border-bottom:1px solid var(--line);text-align:left}
.pi-table thead th{background:#f1f5f9;color:#334155}
.pi-badge{display:inline-block;padding:4px 8px;border-radius:8px;font-size:12px;font-weight:600}
.badge-urgent{background:#fee2e2;color:var(--bad)}
.badge-auto{background:#dcfce7;color:var(--good)}
.badge-obs{background:#fef9c3;color:#854d0e}
.pi-pre{white-space:pre-wrap;background:#f8fafc;border:1px solid var(--line);border-radius:12px;padding:10px;font-size:12px}
.pi-skeleton{height:36px;background:linear-gradient(90deg,#f1f5f9, #e2e8f0, #f1f5f9);background-size:200% 100%;animation:shimmer 1.2s infinite}
@keyframes shimmer{0%{background-position:200% 0}100%{background-position:-200% 0}}
`;

const SEVERITIES = [
    { value: "leve", label: "Leve" },
    { value: "moderado", label: "Moderado" },
    { value: "severo", label: "Severo" },
];

export default function PatientIntakeAndDiagnosis() {
    // Inyectar CSS una vez
    useEffect(() => {
        if (!document.getElementById(STYLE_TAG_ID)) {
            const tag = document.createElement("style");
            tag.id = STYLE_TAG_ID;
            tag.innerHTML = BASE_CSS;
            document.head.appendChild(tag);
        }
    }, []);

    // Catálogos desde backend
    const [symptoms, setSymptoms] = useState([]);
    const [medications, setMedications] = useState([]);
    const [chronicConditions, setChronicConditions] = useState([]);

    // Selecciones
    const [selectedSymptoms, setSelectedSymptoms] = useState({});
    const [selectedAllergies, setSelectedAllergies] = useState([]);
    const [selectedChronics, setSelectedChronics] = useState([]);

    // Estado
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null); // mensaje amigable
    const [results, setResults] = useState([]);
    const [generatedAt, setGeneratedAt] = useState(null);
    const [rulesGlobal, setRulesGlobal] = useState("");

    // Historial en sessionStorage
    const HISTORY_KEY = "DIAG_HISTORY";
    const [history, setHistory] = useState([]);

    // Utilidad: intenta parsear JSON de forma segura
    const safeJson = async (res) => {
        const ctype = res.headers.get("content-type") || "";
        if (!ctype.includes("application/json")) {
            // Devolver vacío si el backend retornó HTML o texto
            return null;
        }
        try { return await res.json(); } catch { return null; }
    };

    useEffect(() => {
        const loadAll = async () => {
            setError(null);
            try {
                const [s, m, c] = await Promise.all([
                    fetch("/api/symptoms", { cache: "no-store" }),
                    fetch("/api/medications", { cache: "no-store" }),
                    fetch("/api/chronic-conditions", { cache: "no-store" }),
                ]);

                const [symData, medData, chrData] = await Promise.all([
                    safeJson(s), safeJson(m), safeJson(c)
                ]);

                setSymptoms(Array.isArray(symData) ? symData : []);
                setMedications(Array.isArray(medData) ? medData : []);
                setChronicConditions(Array.isArray(chrData) ? chrData : []);

                // Si algún catálogo no llegó en JSON, mostrar aviso general (no técnico)
                if (!Array.isArray(symData) || !Array.isArray(medData) || !Array.isArray(chrData)) {
                    setError("El servicio no está disponible temporalmente. Inténtalo más tarde.");
                }
            } catch {
                setError("No fue posible cargar los catálogos en este momento.");
            }
        };

        // Historial de sesión
        try {
            const raw = sessionStorage.getItem(HISTORY_KEY);
            setHistory(raw ? JSON.parse(raw) : []);
        } catch { setHistory([]); }

        loadAll();
    }, []);

    useEffect(() => {
        try { sessionStorage.setItem(HISTORY_KEY, JSON.stringify(history)); } catch {}
    }, [history]);

    const toggleSymptom = (id) => {
        setSelectedSymptoms((prev) => {
            const next = { ...prev };
            if (next[id]) delete next[id]; else next[id] = "leve";
            return next;
        });
    };
    const changeSeverity = (id, severity) => setSelectedSymptoms((p) => ({ ...p, [id]: severity }));
    const toggleFromList = (value, list, setter) => setter(list.includes(value) ? list.filter(v => v!==value) : [...list, value]);

    const selectedSymptomArray = useMemo(() => Object.entries(selectedSymptoms).map(([id, severity]) => ({ id, severity })), [selectedSymptoms]);
    const canAnalyze = selectedSymptomArray.length > 0 && !loading;

    const handleAnalyze = async () => {
        setError(null);
        if (!canAnalyze) { setError("Selecciona al menos un síntoma para analizar."); return; }
        setLoading(true);
        try {
            const body = {
                symptoms: selectedSymptomArray.map((s) => ({ id: isNaN(+s.id) ? s.id : +s.id, severity: s.severity })),
                allergies: selectedAllergies.map((id) => (isNaN(+id) ? id : +id)),
                chronicConditions: selectedChronics.map((id) => (isNaN(+id) ? id : +id)),
            };
            const res = await fetch("/api/diagnosis", { method:"POST", headers:{"Content-Type":"application/json"}, body: JSON.stringify(body) });
            const data = await safeJson(res);
            if (!res.ok || !data) throw new Error();

            const sorted = (data.results || []).slice().sort((a,b)=> (b.affinity||0)-(a.affinity||0));
            setResults(sorted);
            setGeneratedAt(data.generatedAt || new Date().toISOString());
            setRulesGlobal(data.rulesGlobal || "");

            const entry = { ts:new Date().toISOString(), input: body, summary: sorted.map(r=>({disease:r.diseaseName, affinity:r.affinity, urgency:r.urgency})).slice(0,5) };
            setHistory((h)=> [entry, ...h].slice(0,5));
        } catch {
            setError("No fue posible completar el análisis en este momento.");
        } finally { setLoading(false); }
    };

    const affinitiesData = useMemo(() => (results||[]).map(r=>({ name:r.diseaseName, affinity: Math.round((r.affinity||0)*100) })), [results]);
    const urgencyBadge = (uRaw) => {
        const u = (uRaw||"").toLowerCase();
        if (u.includes("inmedi")) return <span className="pi-badge badge-urgent">Consulta médica inmediata sugerida</span>;
        if (u.includes("auto")) return <span className="pi-badge badge-auto">Posible automanejo</span>;
        return <span className="pi-badge badge-obs">Observación recomendada</span>;
    };

    const downloadPDF = async () => {
        const node = document.getElementById("diagnostic-report");
        if (!node) return;
        const canvas = await html2canvas(node, { scale: 2, useCORS: true });
        const imgData = canvas.toDataURL("image/png");
        const pdf = new jsPDF({ orientation: "p", unit: "pt", format: "a4" });
        const pageWidth = pdf.internal.pageSize.getWidth();
        const pageHeight = pdf.internal.pageSize.getHeight();
        const imgWidth = pageWidth - 40;
        const imgHeight = (canvas.height * imgWidth) / canvas.width;
        let y = 20;
        pdf.setFontSize(14); pdf.text("Informe de Análisis Clínico", 20, y); y += 18;
        pdf.setFontSize(10); pdf.text(`Fecha: ${new Date(generatedAt || Date.now()).toLocaleString()}`, 20, y); y += 10;
        pdf.addImage(imgData, "PNG", 20, y, imgWidth, Math.min(imgHeight, pageHeight - y - 20));
        pdf.save(`informe-diagnostico-${new Date().toISOString().slice(0,10)}.pdf`);
    };

    return (
        <div className="pi-wrap">
            <div className="pi-container">
                <header className="pi-header">
                    <h1>Módulo de Pacientes: Ingreso y Análisis</h1>
                    <p className="pi-muted">Selecciona síntomas y severidad, registra alergias y condiciones crónicas. Luego solicita el análisis basado en reglas.</p>
                </header>

                {error && <div className="pi-banner">{error}</div>}

                <section className="pi-grid">
                    {/* Síntomas */}
                    <div className="pi-card">
                        <h2 className="pi-title">Síntomas presentes</h2>
                        {symptoms.length === 0 ? (
                            <div className="pi-skeleton" />
                        ) : (
                            <div className="pi-symptoms">
                                {symptoms.map((s) => {
                                    const checked = selectedSymptoms[s.id] != null;
                                    return (
                                        <div key={s.id} className={`pi-chip ${checked ? "active" : ""}`}>
                                            <label>
                                                <input type="checkbox" checked={checked} onChange={() => toggleSymptom(s.id)} className="mt-1"/>
                                                <div>
                                                    <div className="font-medium">{s.name}</div>
                                                    {s.description && <div className="desc">{s.description}</div>}
                                                </div>
                                            </label>
                                            {checked && (
                                                <div className="pi-row">
                                                    <span className="pi-muted" style={{fontSize:12}}>Severidad:</span>
                                                    <select value={selectedSymptoms[s.id]} onChange={(e)=>changeSeverity(s.id, e.target.value)} className="pi-select">
                                                        {SEVERITIES.map((sv)=>(<option key={sv.value} value={sv.value}>{sv.label}</option>))}
                                                    </select>
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>

                    {/* Alergias / Crónicas */}
                    <div className="pi-card">
                        <h2 className="pi-title">Alergias y Condiciones</h2>

                        <div style={{marginBottom:16}}>
                            <div className="pi-muted" style={{fontSize:13, marginBottom:6}}>Alergias a medicamentos</div>
                            <div className="pi-list">
                                {medications.length === 0 ? (
                                    <div className="pi-skeleton" />
                                ) : (
                                    medications.map((m) => (
                                        <label key={m.id} className="pi-row" style={{padding:"4px 0", cursor:"pointer"}}>
                                            <input type="checkbox" checked={selectedAllergies.includes(m.id)} onChange={()=>toggleFromList(m.id, selectedAllergies, setSelectedAllergies)} />
                                            <span style={{fontSize:14}}>{m.name}</span>
                                        </label>
                                    ))
                                )}
                            </div>
                        </div>

                        <div>
                            <div className="pi-muted" style={{fontSize:13, marginBottom:6}}>Enfermedades crónicas</div>
                            <div className="pi-list">
                                {chronicConditions.length === 0 ? (
                                    <div className="pi-skeleton" />
                                ) : (
                                    chronicConditions.map((c) => (
                                        <label key={c.id} className="pi-row" style={{padding:"4px 0", cursor:"pointer"}}>
                                            <input type="checkbox" checked={selectedChronics.includes(c.id)} onChange={()=>toggleFromList(c.id, selectedChronics, setSelectedChronics)} />
                                            <span style={{fontSize:14}}>{c.name}</span>
                                        </label>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </section>

                <div className="pi-actions">
                    <button onClick={handleAnalyze} disabled={!canAnalyze} className="pi-btn primary">{loading ? "Analizando…" : "Solicitar análisis"}</button>
                    <button onClick={()=>{ setSelectedSymptoms({}); setSelectedAllergies([]); setSelectedChronics([]); setResults([]); setRulesGlobal(""); setGeneratedAt(null); }} className="pi-btn secondary">Limpiar</button>
                    {results.length > 0 && (
                        <button onClick={downloadPDF} className="pi-btn success">Descargar informe PDF</button>
                    )}
                </div>

                <section id="diagnostic-report" className="pi-card">
                    <div style={{display:"flex",gap:16,alignItems:"center",marginBottom:8}}>
                        <h2 className="pi-title" style={{margin:0}}>Resultados del análisis</h2>
                        {generatedAt && <div className="pi-muted" style={{fontSize:12}}>Generado: {new Date(generatedAt).toLocaleString()}</div>}
                    </div>

                    {results.length === 0 ? (
                        <p className="pi-muted">Aún no hay resultados. Completa el formulario y solicita el análisis.</p>
                    ) : (
                        <div style={{display:"grid", gap:24}}>
                            <div style={{overflowX:"auto"}}>
                                <table className="pi-table">
                                    <thead>
                                    <tr>
                                        <th>Enfermedad</th>
                                        <th>Afinidad</th>
                                        <th>Medicamento sugerido</th>
                                        <th>Urgencia</th>
                                        <th>Advertencias</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {results.map((r, i) => (
                                        <tr key={i}>
                                            <td style={{fontWeight:600}}>{r.diseaseName}</td>
                                            <td>{Math.round((r.affinity||0)*100)}%</td>
                                            <td>{r?.medication?.name || "—"}</td>
                                            <td>{urgencyBadge(r?.urgency)}</td>
                                            <td>
                                                {(r?.conflicts?.length>0) ? (
                                                    <span className="pi-badge" style={{background:"#fef2f2", color:"#991b1b"}}>{r.conflicts.join(", ")}</span>
                                                ) : (
                                                    <span className="pi-muted" style={{fontSize:12}}>Sin advertencias</span>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>

                            <div>
                                <h3 style={{fontSize:14, margin:"0 0 8px"}}>Barras de afinidad</h3>
                                <div style={{height:280, width:"100%"}}>
                                    <ResponsiveContainer width="100%" height="100%">
                                        <BarChart data={affinitiesData} margin={{ top: 10, right: 20, left: 0, bottom: 20 }}>
                                            <CartesianGrid strokeDasharray="3 3" />
                                            <XAxis dataKey="name" angle={-15} textAnchor="end" interval={0} height={60} />
                                            <YAxis unit="%" domain={[0, 100]} />
                                            <Tooltip formatter={(v) => `${v}%`} />
                                            <Bar dataKey="affinity" />
                                        </BarChart>
                                    </ResponsiveContainer>
                                </div>
                            </div>

                            <div>
                                <h3 style={{fontSize:14, margin:"0 0 8px"}}>Explicación (reglas activadas)</h3>
                                {rulesGlobal ? (
                                    <pre className="pi-pre">{rulesGlobal}</pre>
                                ) : (
                                    <p className="pi-muted">El sistema no devolvió detalle de reglas activadas.</p>
                                )}
                            </div>
                        </div>
                    )}
                </section>

                <section className="pi-card" style={{marginTop:24}}>
                    <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:8}}>
                        <h2 className="pi-title" style={{margin:0}}>Historial de esta sesión</h2>
                        <button onClick={()=>setHistory([])} className="pi-btn">Limpiar historial</button>
                    </div>
                    {history.length === 0 ? (
                        <p className="pi-muted">No hay análisis previos en esta sesión.</p>
                    ) : (
                        <div style={{display:"grid",gap:12}}>
                            {history.map((h,i)=> (
                                <div key={i} className="pi-chip">
                                    <div className="pi-muted" style={{fontSize:12}}>{new Date(h.ts).toLocaleString()}</div>
                                    <div style={{fontWeight:600, marginTop:6, fontSize:14}}>Top diagnósticos</div>
                                    <ul style={{margin:"4px 0 0 18px"}}>
                                        {h.summary.map((s,idx)=> (
                                            <li key={idx} style={{fontSize:14}}>
                                                {s.disease} — {Math.round((s.affinity||0)*100)}% — {s.urgency || ""}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            ))}
                        </div>
                    )}
                </section>
            </div>
        </div>
    );
}
