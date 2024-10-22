import React, { useState } from 'react';
import './Entrada.css'; // Asegúrate de tener este archivo CSS

export const Entrada = ({ onInputChange }) => {
    const [inputText, setInputText] = useState('');

    return (
        <div className="entrada-container">
            <h2 className="entrada-title">Entrada</h2>
            <textarea 
                className="entrada-textarea"
                value={inputText} 
                onChange={(e) => setInputText(e.target.value)}
                placeholder="Escribe tu texto aquí..."
                rows="10"
                cols="400"  // Ajusta el ancho si es necesario
            />
        </div>
    );
};