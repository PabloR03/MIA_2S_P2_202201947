import React from 'react';
import './Salida.css'; // Asegúrate de tener este archivo CSS

export const Salida = ({ outputText }) => {
    return (
        <div className="salida-container">
            <h2 className="salida-title">Salida</h2>
            <textarea
                className="salida-textarea"
                value={outputText}
                readOnly
                rows="10"
                cols="400" // Aumenta el número de columnas para igualar el ancho
            />
        </div>
    );
};
