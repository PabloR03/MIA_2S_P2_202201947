import './App.css';
import '@fortawesome/fontawesome-free/css/all.min.css';
import React, { useEffect, useState } from 'react';
import Swal from "sweetalert2";

function App() {
  const [inputText, setInputText] = useState('');
  const [outputText, setOutputText] = useState('');
  const [usuario, setUsuario] = useState('');

  // Recuperar el usuario desde localStorage
  useEffect(() => {
    const loggedUser = localStorage.getItem('loggedUser');
    setUsuario(loggedUser || 'No Existe Usuario Logueado.');
  }, []);

  // Función para abrir el explorador de archivos
  const triggerFileSelect = () => {
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    fileInput.accept = '.smia';
    fileInput.onchange = handleFileSelect;
    fileInput.click();
  };

  const handleFileSelect = (event) => {
    const fileInput = event.target.files[0];
    if (fileInput && fileInput.name.endsWith('.smia')) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setInputText(e.target.result);
      };
      reader.readAsText(fileInput);
    } else {
      Swal.fire("Se debe seleccionar un archivo con la extensión '.smia'", "Error", "error");
    }
  };

  // Función para enviar el texto al backend "todos los comandos"
  const handleExecute = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: inputText,
      });
      const text = await response.text();
      setOutputText(text);
    } catch (error) {
      Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
      setOutputText('Error al procesar los comandos de entrada.'); 
    }
  };

  // Función para enviar el texto al backend "logout"
  const handleLogout = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: "logout",
      });
      const text = await response.text();

      if (text.includes("Sesión cerrada con éxito de la partición")) {
        Swal.fire(text, "Cerrando Sesión", "success");
        localStorage.removeItem("loggedUser");
        setTimeout(() => {
          window.location.href = '/';
        }, 2000);
      } else {
        Swal.fire(text, "Error Al Cerrar Sesión", "error");
      }
      setOutputText(text);

    } catch (error) {
      Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
      setOutputText('Error al procesar los comandos de entrada.'); 
    }
  };

  return (
    <div className="container">
      <div className="buttons">
        <button id="selectBtn" onClick={triggerFileSelect}>
          <i className="fas fa-folder-open"></i> Seleccionar
        </button>
        <button id="executeBtn" onClick={handleExecute}>
          <i className="fas fa-play"></i> Ejecutar
        </button>

        <button id="logoutBtn" onClick={handleLogout}>
          <i className="fas fa-right-from-bracket"></i> Cerrar Sesión
        </button>
      </div>

      <div className='usuarioLogueado'>
          <h2 className='txtUsuario'>Usuario: {usuario}</h2>
      </div>

      <div className="input-section">
        <h2>Entrada:</h2>
        <textarea id="inputArea" value={inputText} onChange={(e) => setInputText(e.target.value)}/>
      </div>

      <div className="output-section">
        <h2>Salida:</h2>
        <textarea id="outputArea" placeholder="Salida" value={outputText} readOnly/>
      </div>
    </div>
  );
}

export default App;