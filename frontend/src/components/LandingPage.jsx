import TopBar from "./TopBar";
import heroImg from "../assets/hero.jpg";

export default function LandingPage() {
  // estilos inline (auto-contenidos)
  const pageBg = {
    minHeight: "100vh",
    background: "linear-gradient(180deg, #fde7f1 0%, #fff0f7 60%, #fff 100%)",
  };

  const section = {
    maxWidth: 1200,
    margin: "24px auto",
    padding: "0 20px",
    boxSizing: "border-box",
  };

  // grid responsivo: 1 columna en pantallas chicas, 2 en grandes automáticamente
  const grid = {
    display: "grid",
    gap: 28,
    gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))",
    alignItems: "center",
  };

  const media = {
    borderRadius: 20,
    overflow: "hidden",
    boxShadow: "0 14px 36px rgba(219,39,119,.18)",
    background: "#fff",
  };

  const imgStyle = {
    width: "100%",
    height: "100%",
    display: "block",
    objectFit: "cover",
  };

  const copy = { padding: "8px 4px" };

  const title = {
    fontSize: "clamp(2rem, 4vw, 3rem)",
    margin: "0 0 12px",
    fontWeight: 800,
    background: "linear-gradient(90deg, #ec4899, #f472b6, #fb7185)",
    WebkitBackgroundClip: "text",
    backgroundClip: "text",
    color: "transparent",
    filter: "drop-shadow(0 8px 18px rgba(236,72,153,.15))",
  };

  const lead = {
    fontSize: "1.08rem",
    lineHeight: 1.75,
    color: "#334155",
    margin: "6px 0 16px",
  };

  const disclaimer = {
    background: "#fff0f7",
    border: "1px solid #f9a8d4",
    color: "#9d174d",
    padding: "12px 14px",
    borderRadius: 12,
    fontSize: ".95rem",
  };

  return (
    <div style={pageBg}>
      <TopBar />

      <main style={{ padding: 24 }}>
        <section style={section}>
          <div style={grid}>
            {/* Columna 1: imagen local */}
            <div style={media}>
              <img src={heroImg} alt="MediLogic - asistencia digital en salud" style={imgStyle} />
            </div>

            {/* Columna 2: texto */}
            <div style={copy}>
              <h2 style={title}>MediLogic</h2>

              <p style={lead}>
                <br />
                MediLogic es un sistema experto diseñado para brindar apoyo diagnóstico preliminar a los usuarios. 
                Su funcionamiento se basa en lógica computacional capaz de analizar síntomas, alergias y enfermedades 
                preexistentes con el fin de generar un informe que incluya posibles diagnósticos, el grado de afinidad con 
                cada uno y medicamentos sugeridos, procurando siempre evitar contraindicaciones.
                <br />
                
                El sistema está construido sobre un motor lógico en Prolog, donde un conjunto de reglas definidas en archivos 
                .pl permite realizar inferencias a partir de los datos ingresados. De esta manera, simula parcialmente el razonamiento 
                de un especialista en salud, ofreciendo orientación médica inicial y material educativo que ayude al usuario a 
                comprender mejor su estado de salud.
                <br />
              </p>

              <div style={disclaimer} role="alert">
                ⚠️ Diseñada para orientar y educar, pero <strong>no sustituye la consulta médica profesional</strong>.
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
