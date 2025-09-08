import { Routes, Route, Navigate } from "react-router-dom";
import LandingPage from "./components/LandingPage.jsx";
import AdminPage   from "./components/AdminPage.jsx";
import UserPage    from "./components/UserPage.jsx";
import LoginAdmin from "./components/LoginAdmin.jsx";


function Dummy({ title }) {
  return (
    <div style={{ padding: 24 }}>
      <h1>{title}</h1>
      <p>Página de ejemplo.</p>
    </div>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/login/admin" element={<LoginAdmin />} />
      <Route path="/admin/page" element={<Dummy title="Administrador" />} />
      <Route path="/diagnostico" element={<Dummy title="Paciente" />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
