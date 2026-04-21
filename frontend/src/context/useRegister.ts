import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { registerWithEmail, registerWithGoogle as googleRedirect, RegisterData } from '../services/auth_service';

export const useRegister = () => {
  const [error, setError] = useState<string>('');
  const navigate = useNavigate();

  const register = async (data: RegisterData): Promise<void> => {
    setError('');

    if (!data.firstName || !data.lastName) {
      setError('Nombre y apellido son obligatorios.');
      return;
    }
    if (data.password !== data.confirmPassword) {
      setError('Las contraseñas no coinciden.');
      return;
    }

    try {
      await registerWithEmail(data);
      navigate('/login');
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const registerWithGoogle = async (): Promise<void> => {
    try {
      setError('');
      await googleRedirect();
    } catch (err) {
      setError((err as Error).message);
    }
  };

  return { register, registerWithGoogle, error };
};