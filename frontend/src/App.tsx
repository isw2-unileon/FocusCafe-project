import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ProtectedRoute } from '@/components/ProtectedRoute';


import Home from '@/pages/Home';
//import Login from '@/pages/Login';
//import Register from '@/pages/Register';
//import Dashboard from '@/pages/Dashboard';
//import StudySession from '@/pages/StudySession';
//import AdminDashboard from '@/pages/AdminDashboard';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/*Public routes*/}
        
        {/*<Route path="/" element={<Login />} />
        <Route path="/register" element={<Register />} />*/}
        
        {/*Private routes*/}
          <Route element={<ProtectedRoute />}>
            <Route path="/home" element={<Home />} />
            {/*<Route path="/dashboard" element={<Dashboard />} />
            <Route path="/studySession" element={<StudySession />} />
            <Route path="/adminDashboard" element={<AdminDashboard />} />*/}
          </Route>


        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  );
}