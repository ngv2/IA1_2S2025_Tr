import { useNavigate } from "react-router-dom";
import { useRef, useState } from "react";
import "./Diagnostico.css";

export default function UserPage() {
  const navigate = useNavigate();

  // refs y estado (simulación local; conecta a tu backend cuando lo tengas listo)
  const fileRef = useRef(null);
  const [output, setOutput] = useState("");

  // Botón volver: intenta ir atrás; si no hay historial, va a "/"
  const handleBack = () => {
    if (window.history.length > 1) navigate(-1);
    else navigate("/");
  };

  const runQuery = () => {
    const query = document.getElementById("queryInput")?.value || "";
    if (!query.trim()) return;
    // TODO: llamar a tu backend para evaluar la consulta
    setOutput((prev) => prev + `\n> Consulta: ${query}\nResultado: (simulado)\n`);
  };

  const addFact = () => {
    const fact = document.getElementById("factInput")?.value || "";
    if (!fact.trim()) return;
    // TODO: enviar “hecho” a tu backend
    setOutput((prev) => prev + `\n> Hecho agregado: ${fact} (simulado)\n`);
  };

  const downloadCode = () => {
    // TODO: descargar desde tu backend; por ahora un ejemplo simple
    const blob = new Blob(["% Hechos/Reglas Prolog (simulado)\nhecho(a).\n"], {
      type: "text/plain",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "codigo.pl";
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleFileChange = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    // TODO: enviar el .pl a tu backend (FormData)
    setOutput((prev) => prev + `\n> Cargado archivo: ${file.name} (simulado)\n`);
  };

  

  return (
    <div className="plg-page">
      {/* Topbar */}
      <div className="plg-topbar">
        <button className="plg-btn-back" onClick={handleBack}>
          ← Volver
        </button>
      </div>

      {/* Contenedor principal */}
      <div className="plg-container">
        <h1>MediLogic</h1>

        <section className="plg-section">
          <h2>Cargar archivo Prolog</h2>
          <label htmlFor="fileInput">Selecciona un archivo .pl:</label>
          <input
            type="file"
            id="fileInput"
            accept=".pl"
            ref={fileRef}
            onChange={handleFileChange}
          />
        </section>

        <section className="plg-section">
          <h2>Consulta</h2>
          <input
            type="text"
            id="queryInput"
            placeholder="Escribe tu consulta Prolog aquí..."
          />
          <button onClick={runQuery}>Ejecutar Consulta</button>
        </section>

        <section className="plg-section">
          <h2>Agregar Hecho</h2>
          <input
            type="text"
            id="factInput"
            placeholder="Escribe el hecho Prolog aquí..."
          />
          <button onClick={addFact}>Agregar Hecho</button>
        </section>

        <section className="plg-section">
          <h2>Descargar Código</h2>
          <button onClick={downloadCode}>Descargar Código Prolog</button>
        </section>

        <textarea
          id="output"
          disabled
          value={output}
          placeholder="Aquí aparecerán los resultados..."
          readOnly
        />
      </div>
    </div>
  );
}

