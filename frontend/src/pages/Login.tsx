import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Coffee, LogIn } from 'lucide-react';

export default function Login() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();


  return (
    <div className="min-h-screen bg-orange-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-[2.5rem] shadow-2xl p-10 border-b-8 border-orange-200">

        <div className="text-center mb-8">
          <div className="bg-orange-100 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-4">
            <Coffee size={40} className="text-orange-600" />
          </div>
          <h1 className="text-4xl font-black text-orange-900 tracking-tight">FocusCafe</h1>
          <p className="text-orange-600 font-medium mt-1">Donde el estudio rinde frutos</p>
        </div>

        <div className="space-y-4">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
          />
          <input
            type="password"
            placeholder="Contraseña"
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
          />
          <button
            type="button"
           // onClick={handleLogin}
            className="w-full bg-orange-600 text-white font-bold py-4 rounded-2xl shadow-lg hover:bg-orange-700 transition-all flex items-center justify-center gap-2"
          >
            <LogIn size={20} />
            Entrar al Café
          </button>
          <Link
            to="/register"
            className="block w-full text-center text-orange-800 text-sm font-bold opacity-60 hover:opacity-100 transition-opacity"
          >
            ¿Nuevo aquí? Regístrate
          </Link>
        </div>

      </div>
    </div>
  );
}