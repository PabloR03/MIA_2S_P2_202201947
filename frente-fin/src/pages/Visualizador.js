import React, { useState, useEffect } from "react";
import Swal from "sweetalert2";
import "./Visualizador.css";

const Visualizador = () => {
  const [RutasDiscos, setRutas] = useState([]); 
  const [Particiones, setParticiones] = useState([]);
  const [DiscoSeleccionado, setDiscoSeleccionado] = useState("");

  // Función para obtener las RutasDiscos de los discos
  const handleGetPathDisk = async () => {
    try {
      const response = await fetch('http://localhost:8080/AnalizadorGo/ProcesarComando', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: "obtenerdiscosruta",
      });

      const text = await response.text();

      if (text === "") {
        Swal.fire("No Existen Discos Creados ", "Visualizador", "error");
      } else {
        const rutasObtenidas = text.split('\n').filter(ruta => ruta.trim() !== "");
        setRutas(rutasObtenidas);
        Swal.fire("Mostrando Discos", "Visualizador", "success");
      }

    } catch (error) {
      Swal.fire("Error Al Enviar Texto: "+error, "Error Desconocido", "error");
    }
  };

  useEffect(() => {
    handleGetPathDisk();
  }, []);

  // Función para obtener las particiones de un disco específico
  const fetchPartitions = (RutaDisco) => {
    setDiscoSeleccionado(getNombreDisco(RutaDisco)); 

    fetch("http://localhost:8080/AnalizadorGo/ObtenerParticiones", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ path: RutaDisco }), 
    })
      .then((response) => response.json())
      .then((data) => {
        setParticiones(data || []); 
      })
      .catch((error) => {
        setParticiones([]); 
      });
  };

  // Función para obtener el nombre del disco a partir de la ruta
  const getNombreDisco = (RutaDisco) => {
    const CadenaRuta = RutaDisco.split('/');
    return CadenaRuta[CadenaRuta.length - 1];
  };

  return (
    <div className="unique-container">
      <h2>Discos Creados</h2>
      <div className="unique-disks">

        {RutasDiscos.length > 0 ? (
          RutasDiscos.map((disk, index) => (
            <div key={index} className="unique-disk-item" onClick={() => fetchPartitions(disk)}>
              <i className="fas fa-hdd unique-disk-icon"></i> 
              <h5>Disco: {getNombreDisco(disk)}</h5>
            </div>
          ))):(<p>No Se Han Creado Discos.</p>)}
      </div>          

      {DiscoSeleccionado && (
        <div>
          <h3>Particiones del Disco: {DiscoSeleccionado}</h3>
          {Particiones && Particiones.length > 0 ? (
            <ul className="unique-partition-list">
              {Particiones.map((partition, index) => (
                <div key={index} className="unique-partition-item">
                <i className="fas fa-server unique-partition-icon"></i> 
                <div className="unique-partition-details">
                  <ul>Nombre: {partition.name}</ul>
                  <ul>Tipo: {partition.type}</ul>
                  <ul>Tamaño: {partition.size}</ul>
                  <ul>Inicio: {partition.start}</ul>
                </div>
              </div>
              ))}
            </ul>
          ):(<p>No Existen Particiones Para Este Disco.</p>)}
        </div>
      )}
    </div>
  );
};

export default Visualizador;
