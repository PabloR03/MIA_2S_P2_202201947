import React, { useState } from 'react';
import './App.css';
import '@fortawesome/fontawesome-free/css/all.min.css';

function App() {
    const [inputText, setInputText] = useState('');  // Almacena el texto de entrada
    const [outputText, setOutputText] = useState('');  // Almacena el texto de salida

    const handleFileSelect = (event) => {
        const fileInput = event.target.files[0];
        if (fileInput && fileInput.name.endsWith('.smia')) {
            const reader = new FileReader();
            reader.onload = (e) => {
                setInputText(e.target.result);
            };
            reader.readAsText(fileInput);
        } else {
            alert('Por favor, seleccione un archivo con la extensión .smia');
        }
    };

    const triggerFileSelect = () => {
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = '.smia';
        fileInput.onchange = handleFileSelect;
        fileInput.click();
    };

    const handleInputChange = (e) => {
        setInputText(e.target.value);  // Actualiza el texto de entrada
    };

    const handleExecute = async () => {
        try {
            const response = await fetch('http://localhost:8080/analizar', {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
                },
                body: inputText,
            });
            const text = await response.text();
            setOutputText(text);
        } catch (error) {
            console.error('Error al enviar el texto:', error);
            setOutputText('Error al enviar el texto'); 
        }
    };

    return (
        <div className="App">
            <header className="App-header">
                <h1>MIA Project 202201947</h1>

                <div className="text-container">
                    <div className="entrada-container">
                        <h2 className="entrada-title">Entrada</h2>
                        <textarea 
                            className="entrada-textarea"
                            value={inputText} 
                            onChange={handleInputChange}
                            placeholder="Escribe tu texto aquí..."
                            rows="100"
                            cols="100"  // Ajusta el ancho si es necesario
                        />
                    </div>

                    <div className="salida-container">
                        <h2 className="salida-title">Salida</h2>
                        <textarea
                            className="salida-textarea"
                            value={outputText}
                            readOnly
                            rows="100"
                            cols="50" // Ajusta el ancho si es necesario
                        />
                    </div>
                </div>

                <div className="navbar">
                    <button className="select" onClick={triggerFileSelect}>
                        <i className="fas fa-file-upload"></i> Cargar Archivo
                    </button>
                    <button className="execute" onClick={handleExecute}>
                        <i className="fas fa-play"></i> Ejecutar
                    </button>
                </div>
            </header>
        </div>
    );
}

export default App;
