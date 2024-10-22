import React from 'react';
import './BarraNavegacion.css';

export const BarraNavegacion = ({ onExecute, triggerFileSelect }) => {
    return (
        <div className="navbar">
            <button className="select" onClick={triggerFileSelect}>
                <i className="fas fa-file-upload"></i> Cargar Archivo
            </button>
            <button className="execute" onClick={onExecute}>
                <i className="fas fa-play"></i> Ejecutar
            </button>
        </div>
    );
};
