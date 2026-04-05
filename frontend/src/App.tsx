import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ProtectedRoute } from '@/components/ProtectedRoute';


import Home from '@/pages/Home';
import { GameProvider } from "./context/GameProvider";
//import Login from '@/pages/Login';
//import Register from '@/pages/Register';
//import Dashboard from '@/pages/Dashboard';
//import StudySession from '@/pages/StudySession';
//import AdminDashboard from '@/pages/AdminDashboard';

export default function App() {
  return (
    <GameProvider>
      <BrowserRouter>
        <Routes>
          {/*Public routes*/}
          <Route path="/" element={<Navigate to="/home" />} />
          {/*<Route path="/" element={<Login />} />
          <Route path="/register" element={<Register />} />*/}
          
          {/*Private routes*/}
            <Route element={<ProtectedRoute />}>
              <Route path="/home" element={<Home />} />
              {/*<Route path="/dashboard" element={<Dashboard />} />
              <Route path="/study" element={<StudySession />} />
              <Route path="/adminDashboard" element={<AdminDashboard />} />*/}
            </Route>


          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </BrowserRouter>
    </GameProvider>
  );
}