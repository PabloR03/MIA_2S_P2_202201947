import React from "react";
import { Route, Routes } from "react-router-dom";
import NavBar from "./components/NavBar";
import App from "./pages/App";
import Login from "./pages/Login";
import Visualizador from "./pages/Visualizador";

const Main = () => {
    return (
        <>
            <NavBar />
            <Routes>
                <Route path="/" element={<App />} />
                <Route path="/execution" element={<App />} />
                *<Route path="/login" element={<Login />} />
                <Route path="/visualizador" element={<Visualizador />} />
            </Routes>
        </>
    );
};

export default Main;
