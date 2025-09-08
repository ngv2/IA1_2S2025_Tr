import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./LoginAdmin.css"; // ⬅️ Nuevo archivo CSS para estilos

export default function LoginAdmin() {
  const navigate = useNavigate();
  const [usuario, setUsuario] = useState("");
  const [contrasena, setContrasena] = useState("");
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e) => {
    e.preventDefault();
    setError(null);

    if (!usuario || !contrasena) {
      setError("Por favor completa usuario y contraseña.");
      return;
    }

    try {
      setLoading(true);
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/users/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ usuario, contrasena }),
      });

      const data = await response.json();

      if (response.ok && data?.OK) {
        localStorage.setItem("access_token", data.RESPUESTA.ACCESS_TOKEN);

        if (data.RESPUESTA.TIPO_USUARIO === "admin") {
          navigate("/admin/dashboard");
        } else {
          setError("Login exitoso, pero tu cuenta no tiene permisos de administrador.");
        }
      } else {
        setError(data?.DESCRIPCION || "Credenciales inválidas.");
      }
    } catch (err) {
      setError("No se pudo conectar con el servidor. Intenta de nuevo.");
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
          <label htmlFor="usuario">Correo o Usuario</label>
          <input
            type="text"
            id="usuario"
            value={usuario}
            onChange={(e) => setUsuario(e.target.value)}
            placeholder="admin@ejemplo.com"
            autoComplete="username"
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
