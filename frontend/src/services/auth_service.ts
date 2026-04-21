const API_URL = 'http://localhost:8081/api';

export const loginWithEmail = async (email: string, password: string): Promise<string> => {
  const res = await fetch(`${API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  const data = await res.json();

  if (!res.ok) throw new Error(data.error || 'Error al iniciar sesión');

  return data.token;
};

export const loginWithGoogle = (): void => {
  window.location.href = `${API_URL}/auth/google`;
};