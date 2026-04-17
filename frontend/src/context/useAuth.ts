import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginWithEmail, loginWithGoogle as googleRedirect } from '../services/auth_service';

export const useAuth = () => {
  const [error, setError] = useState<string>('');
  const navigate = useNavigate();

  const login = async (email: string, password: string): Promise<void> => {
    try {
      setError('');
      const token = await loginWithEmail(email, password);
      localStorage.setItem('token', token);
      navigate('/home');
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const loginWithGoogle = (): void => {
    googleRedirect();
  };

  return { login, loginWithGoogle, error };
};