import React from 'react';
import './NavBar.css';
import { Link } from 'react-router-dom';

const NavBar = () => {
    return (
        <nav className="navbar">
        <div className="logo">
            <h2>f2 202201947</h2>
        </div>
        <ul className="nav-links">
            <li>
            <Link to="/execution">Ejecución</Link>
            </li>
            <li>
            <Link to="/login">Iniciar sesión</Link>
            </li>
            <li>
            <Link to="/visualizador">Visualizador</Link>
            </li>
        </ul>
        </nav>
    );
    };

export default NavBar;
