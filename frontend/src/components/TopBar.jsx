import { useNavigate } from "react-router-dom";

export default function TopBar() {
  const navigate = useNavigate();

  const wrap = {
    position: "sticky",
    top: 0,
    zIndex: 1000,
    width: "100vw",
    marginLeft: "calc(50% - 50vw)",                // ocupa todo el ancho real
    background: "linear-gradient(180deg,#fff 0%, #fff 40%, #fde7f1 100%)",
    padding: "18px 16px",
    boxSizing: "border-box",
  };

  const pill = {
    width: "100%",
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 12,
    padding: "10px 14px",
    background: "#ffffff",
    border: "1px solid #f1c8d7",
    borderRadius: 9999,
    boxShadow: "0 12px 32px rgba(236,72,153,.15)",
    boxSizing: "border-box",
  };

  const brand = {
    display: "inline-flex",
    alignItems: "center",
    gap: 10,
    background: "transparent",
    border: 0,
    cursor: "pointer",
  };

  const logo = {
    width: 36, height: 36, borderRadius: 9999,
    background:
      "radial-gradient(circle at 30% 35%, #fecdd3 0 30%, transparent 31%)," +
      "radial-gradient(circle at 70% 65%, #f9a8d4 0 30%, transparent 31%)," +
      "radial-gradient(circle at 50% 50%, #f472b6 0 46%, transparent 47%)",
    boxShadow: "0 6px 16px rgba(219,39,119,.22)",
  };

  const title = { fontWeight: 700, color: "#0f172a", letterSpacing: ".2px" };

  const actions = { display: "flex", alignItems: "center", gap: 10 };

  const btnBase = {
    appearance: "none",
    textDecoration: "none",
    userSelect: "none",
    padding: "10px 18px",
    borderRadius: 9999,
    fontWeight: 700,
    border: "1px solid transparent",
    cursor: "pointer",
    transition: "transform .12s, box-shadow .12s, background .12s, color .12s, border-color .12s",
  };
  const btnOutline = {
    ...btnBase,
    background: "#fff",
    color: "#0f172a",
    borderColor: "#e5e7eb",
  };
  const btnFill = {
    ...btnBase,
    background: "#ec4899",
    color: "#fff",
    borderColor: "#db2777",
  };

  return (
    <header style={wrap}>
      <div style={pill}>
        <button style={brand} onClick={() => navigate("/")}>
          <span style={logo} aria-hidden />
          <span style={title}>MediLogic</span>
        </button>

        <nav style={actions} aria-label="primary">
          <button
            style={btnOutline}
            onMouseEnter={(e) => { e.currentTarget.style.borderColor = "#ec4899"; e.currentTarget.style.color = "#db2777"; }}
            onMouseLeave={(e) => { e.currentTarget.style.borderColor = "#e5e7eb"; e.currentTarget.style.color = "#0f172a"; }}
            onClick={() => navigate("/login/admin")}
          >
            Login admin
          </button>

          <button
            style={btnFill}
            onMouseEnter={(e) => { e.currentTarget.style.background = "#db2777"; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = "#ec4899"; }}
            onClick={() => navigate("/diagnostico")}
          >
            Paciente
          </button>
        </nav>
      </div>
    </header>
  );
}
