export default function LoginAdmin() {
  return (
    <div>
      <h1>Login Administrativo</h1>
      <form>
        <label htmlFor="usuario">Usuario</label>
        <input type="text" id="usuario" name="usuario" />

        <label htmlFor="contrasena">Contraseña</label>
        <input type="password" id="contrasena" name="contrasena" />

        <button type="submit">Iniciar Sesión</button>
      </form>
    </div>
  );
}
