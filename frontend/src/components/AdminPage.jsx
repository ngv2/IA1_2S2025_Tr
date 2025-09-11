// components/AdminPage.jsx
import { useNavigate } from "react-router-dom";
import { useRef, useState, useEffect } from "react";
import "./Diagnostico.css";

/* Ítem lista con botones a la derecha */
function ListItem({ text, onEdit, onDelete }) {
  return (
    <div className="plg-list-item">
      <span className="plg-list-text">{text}</span>
      <div className="plg-list-actions">
        <button className="plg-chip outline" onClick={onDelete}>eliminar</button>
        <button className="plg-chip" onClick={onEdit}>modificar</button>
      </div>
    </div>
  );
}

export default function AdminPage() {
  const navigate = useNavigate();
  const importRef = useRef(null);

  /* ===== Estado base (simulado) ===== */
  const [diseases, setDiseases] = useState(["Enfermedad1", "Enfermedad2"]);
  const [symptoms, setSymptoms] = useState(["Sintoma1", "Sintoma2"]);
  const [newDisease, setNewDisease] = useState("");
  const [newSymptom, setNewSymptom] = useState("");

  /* ===== Medicamentos ===== */
  // { id, nombre, trata:[enfermedades], contra:string }
  const [meds, setMeds] = useState([
    { id: 1, nombre: "Paracetamol", trata: ["Enfermedad1"], contra: "Hepatopatías severas" },
  ]);
  const [medNombre, setMedNombre] = useState("");
  const [medContra, setMedContra] = useState("");
  const [medTrata, setMedTrata] = useState(new Set());

  /* ===== Catálogos Crónicas / Alergias ===== */
  const [cronicas, setCronicas] = useState(["Diabetes", "InsuficienciaRenal"]);
  const [alergias, setAlergias] = useState(["Penicilina", "Lactosa"]);
  const [newCronica, setNewCronica] = useState("");
  const [newAlergia, setNewAlergia] = useState("");

  /* ===== Relaciones de contraindicación =====
     relations[medId] = { cronicas:[...], alergias:[...] } */
  const [relations, setRelations] = useState({
    1: { cronicas: ["InsuficienciaRenal"], alergias: ["Lactosa"] },
  });
  const [selMedId, setSelMedId] = useState(1); // medicamento seleccionado en el editor de relaciones

  /* ===== Clasificación: sistemas / tipos ===== */
  const [sistemas, setSistemas] = useState(["Respiratorio", "Digestivo", "Endocrino"]);
  const [tipos, setTipos] = useState(["Viral", "Crónico", "Inmunológico"]);
  const [newSistema, setNewSistema] = useState("");
  const [newTipo, setNewTipo] = useState("");

  // Mapa de clasificaciones por enfermedad:
  // clasif["Enfermedad1"] = { sistemas: [...], tipos: [...] }
  const [clasif, setClasif] = useState({});

  // Enfermedad seleccionada para editar sus clasificaciones
  const [selDiseaseForClass, setSelDiseaseForClass] = useState("");

  // mantener selección válida cuando cambian enfermedades
  useEffect(() => {
    if (!diseases.includes(selDiseaseForClass)) {
      setSelDiseaseForClass(diseases[0] || "");
    }
  }, [diseases, selDiseaseForClass]);

  const [output, setOutput] = useState("");

  const handleBack = () => (window.history.length > 1 ? navigate(-1) : navigate("/"));

  /* ===== CRUD Enfermedades ===== */
  const addDisease = () => {
    const v = newDisease.trim();
    if (!v) return;
    setDiseases((arr) => [...arr, v]);
    setNewDisease("");
  };
  const editDisease = (i) => {
    const cur = diseases[i];
    const v = window.prompt("Modificar enfermedad:", cur);
    if (v && v.trim()) {
      const copy = [...diseases];
      copy[i] = v.trim();
      setDiseases(copy);
    }
  };
  const deleteDisease = (i) => setDiseases((arr) => arr.filter((_, idx) => idx !== i));

  /* ===== CRUD Síntomas ===== */
  const addSymptom = () => {
    const v = newSymptom.trim();
    if (!v) return;
    setSymptoms((arr) => [...arr, v]);
    setNewSymptom("");
  };
  const editSymptom = (i) => {
    const cur = symptoms[i];
    const v = window.prompt("Modificar síntoma:", cur);
    if (v && v.trim()) {
      const copy = [...symptoms];
      copy[i] = v.trim();
      setSymptoms(copy);
    }
  };
  const deleteSymptom = (i) => setSymptoms((arr) => arr.filter((_, idx) => idx !== i));

  /* ===== CRUD Medicamentos ===== */
  const toggleTrata = (enf) => {
    setMedTrata((prev) => {
      const n = new Set(prev);
      n.has(enf) ? n.delete(enf) : n.add(enf);
      return n;
    });
  };
  const addMed = () => {
    const nombre = medNombre.trim();
    const contra = medContra.trim();
    const trata = Array.from(medTrata);
    if (!nombre) return;

    const id = (meds.at(-1)?.id || 0) + 1;
    setMeds((arr) => [...arr, { id, nombre, trata, contra }]);
    setRelations((r) => ({ ...r, [id]: { cronicas: [], alergias: [] } }));
    setSelMedId(id);

    setMedNombre(""); setMedContra(""); setMedTrata(new Set());
  };
  const editMed = (id) => {
    const m = meds.find((x) => x.id === id);
    if (!m) return;
    const nombre = window.prompt("Nombre del medicamento:", m.nombre) ?? m.nombre;
    const contra = window.prompt("Contraindicaciones (texto libre):", m.contra ?? "") ?? m.contra;
    setMeds((arr) => arr.map((x) => (x.id === id ? { ...x, nombre: nombre.trim() || x.nombre, contra } : x)));
  };
  const deleteMed = (id) => {
    setMeds((arr) => arr.filter((x) => x.id !== id));
    setRelations((r) => {
      const c = { ...r }; delete c[id]; return c;
    });
    if (selMedId === id) setSelMedId(meds.find((m) => m.id !== id)?.id ?? 0);
  };

  /* ===== Catálogo Crónicas / Alergias ===== */
  const addCronica = () => {
    const v = newCronica.trim();
    if (!v) return; setCronicas((a) => [...a, v]); setNewCronica("");
  };
  const delCronica = (i) => setCronicas((a) => a.filter((_, idx) => idx !== i));

  const addAlergia = () => {
    const v = newAlergia.trim();
    if (!v) return; setAlergias((a) => [...a, v]); setNewAlergia("");
  };
  const delAlergia = (i) => setAlergias((a) => a.filter((_, idx) => idx !== i));

  /* ===== Editor de relaciones ===== */
  const relFor = relations[selMedId] || { cronicas: [], alergias: [] };
  const toggleRel = (kind, value) => {
    setRelations((prev) => {
      const current = prev[selMedId] || { cronicas: [], alergias: [] };
      const set = new Set(current[kind]);
      set.has(value) ? set.delete(value) : set.add(value);
      return { ...prev, [selMedId]: { ...current, [kind]: Array.from(set) } };
    });
  };

  /* ===== Catálogo Sistemas / Tipos ===== */
  const addSistema = () => {
    const v = newSistema.trim();
    if (!v) return;
    if (!sistemas.includes(v)) setSistemas((a) => [...a, v]);
    setNewSistema("");
  };
  const delSistema = (i) => setSistemas((a) => a.filter((_, idx) => idx !== i));

  const addTipo = () => {
    const v = newTipo.trim();
    if (!v) return;
    if (!tipos.includes(v)) setTipos((a) => [...a, v]);
    setNewTipo("");
  };
  const delTipo = (i) => setTipos((a) => a.filter((_, idx) => idx !== i));

  // Toggle de clasificación para la enfermedad seleccionada
  const toggleClasif = (kind /* 'sistemas' | 'tipos' */, value) => {
    const enf = selDiseaseForClass;
    if (!enf) return;
    setClasif((prev) => {
      const cur = prev[enf] || { sistemas: [], tipos: [] };
      const set = new Set(cur[kind]);
      set.has(value) ? set.delete(value) : set.add(value);
      return { ...prev, [enf]: { ...cur, [kind]: Array.from(set) } };
    });
  };

  /* ===== Acciones inferiores / Export ===== */
  const runRPA = () => setOutput((p) => p + "\n> RPA ejecutado (simulado)\n");
  const handleImportClick = () => importRef.current?.click();
  const handleImport = (e) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setOutput((p) => p + `\n> Importado: ${f.name} (simulado)\n`);
    e.target.value = "";
  };

  const handleExport = () => {
    const content = [
      "% Export Prolog simulado",
      "% Enfermedades",
      ...diseases.map((d) => `enfermedad('${d}').`),
      "% Síntomas",
      ...symptoms.map((s) => `sintoma('${s}').`),
      "% Medicamentos",
      ...meds.map(
        (m) =>
          `medicamento('${m.nombre}', [${m.trata.map((t) => `'${t}'`).join(", ")}], '${(m.contra || "").replace(/'/g, "\\'")}').`
      ),
      "% Catalogo cronicas",
      ...cronicas.map((c) => `cronica('${c}').`),
      "% Catalogo alergias",
      ...alergias.map((a) => `alergia('${a}').`),
      "% Relaciones de contraindicación",
      ...Object.entries(relations).flatMap(([mid, rel]) => {
        const m = meds.find((x) => String(x.id) === String(mid));
        if (!m) return [];
        return [
          ...rel.cronicas.map((c) => `contraindicado_si_cronica('${m.nombre}', '${c}').`),
          ...rel.alergias.map((a) => `contraindicado_por_alergia('${m.nombre}', '${a}').`),
        ];
      }),

      "% Catálogo de sistemas",
      ...sistemas.map((s) => `sistema('${s}').`),
      "% Catálogo de tipos",
      ...tipos.map((t) => `tipo('${t}').`),

      "% Clasificación de enfermedades",
      ...Object.entries(clasif).flatMap(([enf, rel]) => ([
        ...(rel?.sistemas || []).map((s) => `enf_sistema('${enf}', '${s}').`),
        ...(rel?.tipos || []).map((t) => `enf_tipo('${enf}', '${t}').`),
      ])),
    ].join("\n");
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "medilogic.pl";
    a.click();
    URL.revokeObjectURL(url);
  };

  const viewPL = () => {
    const preview = [
      "%% Vista rápida (simulada)",
      ...diseases.map((d) => `enfermedad('${d}').`),
      ...symptoms.map((s) => `sintoma('${s}').`),
      ...meds.map(
        (m) =>
          `medicamento('${m.nombre}', [${m.trata.map((t) => `'${t}'`).join(", ")}], '${m.contra || ""}').`
      ),
      ...cronicas.map((c) => `cronica('${c}').`),
      ...alergias.map((a) => `alergia('${a}').`),
      ...Object.entries(relations).flatMap(([mid, rel]) => {
        const m = meds.find((x) => String(x.id) === String(mid));
        if (!m) return [];
        return [
          ...rel.cronicas.map((c) => `contraindicado_si_cronica('${m.nombre}', '${c}').`),
          ...rel.alergias.map((a) => `contraindicado_por_alergia('${m.nombre}', '${a}').`),
        ];
      }),
      ...sistemas.map((s) => `sistema('${s}').`),
      ...tipos.map((t) => `tipo('${t}').`),
      ...Object.entries(clasif).flatMap(([enf, rel]) => ([
        ...(rel?.sistemas || []).map((s) => `enf_sistema('${enf}', '${s}').`),
        ...(rel?.tipos || []).map((t) => `enf_tipo('${enf}', '${t}').`),
      ])),
    ].join("\n");
    setOutput((p) => p + "\n" + preview + "\n");

      const handleLogout = () => {
      // Aquí puedes limpiar tokens/localStorage si los usas
      // localStorage.removeItem("token");
      navigate("/login/admin"); // redirige al login de admin
  };

  };

  /* ===== RENDER ===== */
  return (
    <div className="plg-page">
      <div className="plg-topbar">
        <button className="plg-btn-back" onClick={handleBack}>← Volver</button>
      </div>

      <div className="plg-container">
        <h1>Módulo de administrador — MediLogic</h1>

        {/* Dos columnas: Enfermedades / Síntomas */}
        <div className="plg-grid-2">
          <section className="plg-box">
            <h2>Enfermedades</h2>
            <div className="plg-add-inline">
              <input
                type="text"
                placeholder="Agregar enfermedad"
                value={newDisease}
                onChange={(e) => setNewDisease(e.target.value)}
              />
              <button className="plg-chip" onClick={addDisease}>Agregar</button>
            </div>
            <div className="plg-card-list">
              {diseases.map((d, i) => (
                <ListItem
                  key={`${d}-${i}`}
                  text={d}
                  onEdit={() => editDisease(i)}
                  onDelete={() => deleteDisease(i)}
                />
              ))}
            </div>
          </section>

          <section className="plg-box">
            <h2>Síntomas</h2>
            <div className="plg-add-inline">
              <input
                type="text"
                placeholder="Agregar síntoma"
                value={newSymptom}
                onChange={(e) => setNewSymptom(e.target.value)}
              />
              <button className="plg-chip" onClick={addSymptom}>Agregar</button>
            </div>
            <div className="plg-card-list">
              {symptoms.map((s, i) => (
                <ListItem
                  key={`${s}-${i}`}
                  text={s}
                  onEdit={() => editSymptom(i)}
                  onDelete={() => deleteSymptom(i)}
                />
              ))}
            </div>
          </section>
        </div>

        {/* ===== Medicamentos ===== */}
        <section className="plg-box" style={{ marginTop: "1.5rem" }}>
          <h2>Medicamentos</h2>

          {/* Alta de medicamentos */}
          <div className="plg-form">
            <div className="plg-form-row">
              <label>Nombre</label>
              <input
                type="text"
                placeholder="Ej. Ibuprofeno"
                value={medNombre}
                onChange={(e) => setMedNombre(e.target.value)}
              />
            </div>

            <div className="plg-form-row">
              <label>Trata</label>
              <div className="plg-pills">
                {diseases.length === 0 && <span className="plg-muted">Primero agrega enfermedades.</span>}
                {diseases.map((enf) => (
                  <label key={enf} className={"plg-pill " + (medTrata.has(enf) ? "is-active" : "")}>
                    <input type="checkbox" checked={medTrata.has(enf)} onChange={() => toggleTrata(enf)} />
                    <span>{enf}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="plg-form-row">
              <label>Contraindicaciones</label>
              <textarea
                rows={3}
                placeholder="Ej. Embarazo, úlcera gástrica, insuficiencia renal…"
                value={medContra}
                onChange={(e) => setMedContra(e.target.value)}
              />
            </div>

            <div className="plg-form-actions">
              <button className="plg-chip" onClick={addMed}>Agregar medicamento</button>
            </div>
          </div>

          {/* Lista de medicamentos */}
          <div className="plg-med-list">
            {meds.length === 0 && <div className="plg-empty">Sin medicamentos registrados</div>}
            {meds.map((m) => (
              <div key={m.id} className="plg-med-card">
                <div className="plg-med-head">
                  <div className="plg-med-title">{m.nombre}</div>
                  <div className="plg-list-actions">
                    <button className="plg-chip outline" onClick={() => deleteMed(m.id)}>eliminar</button>
                    <button className="plg-chip" onClick={() => editMed(m.id)}>modificar</button>
                  </div>
                </div>
                <div className="plg-med-body">
                  <div>
                    <div className="plg-label">Trata:</div>
                    <div className="plg-tags">
                      {(m.trata ?? []).map((t) => <span key={t} className="plg-tag">{t}</span>)}
                      {(!m.trata || m.trata.length === 0) && <span className="plg-muted">—</span>}
                    </div>
                  </div>
                  <div>
                    <div className="plg-label">Contra (texto):</div>
                    <div className="plg-text">{m.contra || "—"}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* ===== Relaciones de contraindicación (med ↔ crónicas / alergias) ===== */}
        <section className="plg-box" style={{ marginTop: "1.5rem" }}>
          <h2>Contraindicaciones avanzadas</h2>

          {/* Catálogos rápidos */}
          <div className="plg-grid-2">
            <div>
              <h3 className="plg-subtitle">Enfermedades crónicas</h3>
              <div className="plg-add-inline">
                <input
                  type="text"
                  placeholder="Ej. Hipertension"
                  value={newCronica}
                  onChange={(e) => setNewCronica(e.target.value)}
                />
                <button className="plg-chip" onClick={addCronica}>Agregar</button>
              </div>
              <div className="plg-card-list">
                {cronicas.map((c, i) => (
                  <div key={c} className="plg-list-item">
                    <span className="plg-list-text">{c}</span>
                    <div className="plg-list-actions">
                      <button className="plg-chip outline" onClick={() => delCronica(i)}>eliminar</button>
                    </div>
                  </div>
                ))}
                {cronicas.length === 0 && <div className="plg-empty">Sin crónicas registradas</div>}
              </div>
            </div>

            <div>
              <h3 className="plg-subtitle">Alergias</h3>
              <div className="plg-add-inline">
                <input
                  type="text"
                  placeholder="Ej. Sulfas"
                  value={newAlergia}
                  onChange={(e) => setNewAlergia(e.target.value)}
                />
                <button className="plg-chip" onClick={addAlergia}>Agregar</button>
              </div>
              <div className="plg-card-list">
                {alergias.map((a, i) => (
                  <div key={a} className="plg-list-item">
                    <span className="plg-list-text">{a}</span>
                    <div className="plg-list-actions">
                      <button className="plg-chip outline" onClick={() => delAlergia(i)}>eliminar</button>
                    </div>
                  </div>
                ))}
                {alergias.length === 0 && <div className="plg-empty">Sin alergias registradas</div>}
              </div>
            </div>
          </div>

          {/* Editor de relaciones */}
          <div className="plg-form" style={{ marginTop: "1rem" }}>
            <div className="plg-form-row">
              <label>Medicamento</label>
              <select
                className="plg-select"
                value={selMedId || ""}
                onChange={(e) => setSelMedId(Number(e.target.value))}
              >
                {meds.map((m) => (
                  <option key={m.id} value={m.id}>{m.nombre}</option>
                ))}
              </select>
            </div>

            <div className="plg-form-row">
              <label>Contraindicado si crónica</label>
              <div className="plg-pills">
                {cronicas.map((c) => {
                  const active = relFor.cronicas.includes(c);
                  return (
                    <label key={c} className={"plg-pill " + (active ? "is-active" : "")}>
                      <input
                        type="checkbox"
                        checked={active}
                        onChange={() => toggleRel("cronicas", c)}
                      />
                      <span>{c}</span>
                    </label>
                  );
                })}
              </div>
            </div>

            <div className="plg-form-row">
              <label>Contraindicado por alergia</label>
              <div className="plg-pills">
                {alergias.map((a) => {
                  const active = relFor.alergias.includes(a);
                  return (
                    <label key={a} className={"plg-pill " + (active ? "is-active" : "")}>
                      <input
                        type="checkbox"
                        checked={active}
                        onChange={() => toggleRel("alergias", a)}
                      />
                      <span>{a}</span>
                    </label>
                  );
                })}
              </div>
            </div>
          </div>
        </section>

        {/* ===== Clasificación de enfermedades por sistema / tipo ===== */}
        <section className="plg-box" style={{ marginTop: "1.5rem" }}>
          <h2>Clasificación de enfermedades</h2>

          <div className="plg-grid-2">
            {/* Catálogos */}
            <div>
              <h3 className="plg-subtitle">Catálogo de Sistemas</h3>
              <div className="plg-add-inline">
                <input
                  type="text"
                  placeholder="Ej. Respiratorio"
                  value={newSistema}
                  onChange={(e) => setNewSistema(e.target.value)}
                />
                <button className="plg-chip" onClick={addSistema}>Agregar</button>
              </div>
              <div className="plg-card-list">
                {sistemas.map((s, i) => (
                  <div key={s} className="plg-list-item">
                    <span className="plg-list-text">{s}</span>
                    <div className="plg-list-actions">
                      <button className="plg-chip outline" onClick={() => delSistema(i)}>eliminar</button>
                    </div>
                  </div>
                ))}
                {sistemas.length === 0 && <div className="plg-empty">Sin sistemas registrados</div>}
              </div>

              <h3 className="plg-subtitle" style={{ marginTop: "1rem" }}>Catálogo de Tipos</h3>
              <div className="plg-add-inline">
                <input
                  type="text"
                  placeholder="Ej. Viral"
                  value={newTipo}
                  onChange={(e) => setNewTipo(e.target.value)}
                />
                <button className="plg-chip" onClick={addTipo}>Agregar</button>
              </div>
              <div className="plg-card-list">
                {tipos.map((t, i) => (
                  <div key={t} className="plg-list-item">
                    <span className="plg-list-text">{t}</span>
                    <div className="plg-list-actions">
                      <button className="plg-chip outline" onClick={() => delTipo(i)}>eliminar</button>
                    </div>
                  </div>
                ))}
                {tipos.length === 0 && <div className="plg-empty">Sin tipos registrados</div>}
              </div>
            </div>

            {/* Asignación a enfermedad */}
            <div>
              <h3 className="plg-subtitle">Asignar a una enfermedad</h3>
              <div className="plg-form">
                <div className="plg-form-row">
                  <label>Enfermedad</label>
                  <select
                    className="plg-select"
                    value={selDiseaseForClass || ""}
                    onChange={(e) => setSelDiseaseForClass(e.target.value)}
                  >
                    {(diseases.length ? diseases : [""]).map((d) => (
                      d ? <option key={d} value={d}>{d}</option> : <option key="none" value="">—</option>
                    ))}
                  </select>
                </div>

                <div className="plg-form-row">
                  <label>Sistema(s)</label>
                  <div className="plg-pills">
                    {sistemas.map((s) => {
                      const active = (clasif[selDiseaseForClass]?.sistemas || []).includes(s);
                      return (
                        <label key={s} className={"plg-pill " + (active ? "is-active" : "")}>
                          <input
                            type="checkbox"
                            checked={active}
                            onChange={() => toggleClasif("sistemas", s)}
                          />
                          <span>{s}</span>
                        </label>
                      );
                    })}
                    {sistemas.length === 0 && <span className="plg-muted">No hay sistemas.</span>}
                  </div>
                </div>

                <div className="plg-form-row">
                  <label>Tipo(s)</label>
                  <div className="plg-pills">
                    {tipos.map((t) => {
                      const active = (clasif[selDiseaseForClass]?.tipos || []).includes(t);
                      return (
                        <label key={t} className={"plg-pill " + (active ? "is-active" : "")}>
                          <input
                            type="checkbox"
                            checked={active}
                            onChange={() => toggleClasif("tipos", t)}
                          />
                          <span>{t}</span>
                        </label>
                      );
                    })}
                    {tipos.length === 0 && <span className="plg-muted">No hay tipos.</span>}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Acciones inferiores */}
        <div className="plg-actions-bottom">
          <button className="plg-btn mono" onClick={runRPA}>RPA</button>
            <input ref={importRef} type="file" accept=".pl" hidden onChange={handleImport} />
            <button className="plg-btn outline" onClick={handleImportClick}>Importar .pl</button>
            <button className="plg-btn primary" onClick={handleExport}>Exportar .pl</button>
            <button className="plg-btn soft" onClick={viewPL}>Ver .pl</button>
          </div>

        <textarea
          id="output"
          disabled
          value={output}
          placeholder="Aquí aparecerán mensajes y previsualizaciones..."
          readOnly
        />
      </div>
    </div>
  );
}
