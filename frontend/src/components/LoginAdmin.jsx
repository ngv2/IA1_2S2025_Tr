import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import "./LoginAdmin.css";

export default function LoginAdmin() {
  const navigate = useNavigate();
  const [usuario, setUsuario] = useState("");
  const [contrasena, setContrasena] = useState("");
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [admins, setAdmins] = useState([]);

  useEffect(() => {
    const loadAdmins = async () => {
      try {
        const res = await fetch("/admins.txt", { cache: "no-store" });
        if (!res.ok) throw new Error("No se pudo cargar admins.txt");
        const text = await res.text();
        const parsed = text
            .split("\n")
            .map(l => l.trim())
            .filter(l => l && !l.startsWith("#"))
            .map(l => {
              const parts = l.split(/\s*-\s*/);
              const email = (parts[0] || "").trim().toLowerCase();
              const password = (parts[1] || "").trim();
              return { email, password };
            })
            .filter(a => a.email && a.password);
        setAdmins(parsed);
      } catch (e) {
        setError("Error cargando configuración de administradores.");
      }
    };
    loadAdmins();
  }, []);

  const handleLogin = async (e) => {
    e.preventDefault();
    setError(null);

    if (!usuario || !contrasena) {
      setError("Por favor completa usuario y contraseña.");
      return;
    }

    setLoading(true);
    try {
      const email = usuario.trim().toLowerCase();
      const pass = contrasena;

      const match = admins.find(a => a.email === email && a.password === pass);

      if (match) {
        localStorage.setItem("access_token", "LOCAL_ADMIN_LOGIN");
        localStorage.setItem("user_email", email);
        localStorage.setItem("user_role", "admin");
        navigate("/admin/page");
      } else {
        setError("Credenciales inválidas o usuario no autorizado.");
      }
    } catch {
      setError("Ocurrió un error al validar credenciales.");
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="login-container">
        <div className="login-card">
          <h2 className="login-title">Ingreso Administrativo</h2>
          <p className="login-sub">Acceso exclusivo para personal autorizado</p>

          <form onSubmit={handleLogin} className="login-form">
            <label htmlFor="usuario">Correo</label>
            <input
                type="email"
                id="usuario"
                value={usuario}
                onChange={(e) => setUsuario(e.target.value)}
                placeholder="admin@ejemplo.com"
                autoComplete="username"
                inputMode="email"
            />

            <label htmlFor="contrasena">Contraseña</label>
            <input
                type="password"
                id="contrasena"
                value={contrasena}
                onChange={(e) => setContrasena(e.target.value)}
                placeholder="••••••••"
                autoComplete="current-password"
            />

            {error && <p className="login-error">{error}</p>}

            <button type="submit" disabled={loading}>
              {loading ? "Ingresando…" : "Iniciar Sesión"}
            </button>
          </form>
        </div>
      </div>
  );
}
