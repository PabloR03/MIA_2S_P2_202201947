import React, { useState } from 'react';
import './Login.css';
import '@fortawesome/fontawesome-free/css/all.min.css';
import Swal from "sweetalert2";

const Login = ({ onLogin }) => {
  const [userId, setUserId] = useState('');
  const [password, setPassword] = useState('');
  const [partitionId, setPartitionId] = useState('');

  const handleLogin = async () => {
    const loginCommand = `login -user="${userId}" -pass="${password}" -id="${partitionId}"`;
    try {
        const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
            method: 'POST',
            headers: {
                'Content-Type': 'text/plain',
            },
            body: loginCommand.toLocaleLowerCase(),
        });
        const text = await response.text();
        
        if (text.includes("Inicio exitoso")) {
            Swal.fire(text, "Bienvenido A La aplicación", "success");
            localStorage.setItem("loggedUser", userId);
            if (onLogin) {
              onLogin(userId);
            }

        } else {
            Swal.fire(text, "Error Al Iniciar Sesión", "error");
        }
        
    } catch (error) {
        Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
    }
}


  return (
    <div className="login-container">
      <h2 className="login-title">Login</h2>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-id-badge"></i> ID Partición
        </label>
        <input
          type="text"
          className="login-input"
          placeholder="Ingrese el ID de la partición"
          value={partitionId}
          onChange={(e) => setPartitionId(e.target.value)}
        />
      </div>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-user"></i> Usuario
        </label>
        <input
          type="text"
          className="login-input"
          placeholder="Ingrese su usuario"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
        />
      </div>
      <div className="login-mb">
        <label className="login-label">
          <i className="fas fa-lock"></i> Contraseña
        </label>
        <input
          type="password"
          className="login-input"
          placeholder="Ingrese su contraseña"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      <button className="login-btn" onClick={handleLogin}>
        <i className="fas fa-sign-in-alt"></i> Iniciar sesión
      </button>
    </div>
  );
};

export default Login;
