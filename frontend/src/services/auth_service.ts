import { createClient } from '@supabase/supabase-js';

const API_URL = 'http://localhost:8081/api';

const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL,
  import.meta.env.VITE_SUPABASE_ANON_KEY
);

// ---- LOGIN ----
export const loginWithEmail = async (email: string, password: string): Promise<string> => {
  const res = await fetch(`${API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error logging in');
  return data.token;
};

export const loginWithGoogle = (): void => {
  window.location.href = `${API_URL}/auth/google`;
};


// ---- REGISTER ----
export interface RegisterData {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export const registerWithEmail = async (data: RegisterData): Promise<void> => {
  const res = await fetch(`${API_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      first_name: data.firstName,
      last_name: data.lastName,
      email: data.email,
      password: data.password,
      confirm_password: data.confirmPassword,
    }),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || 'Error during registration');
};

export const registerWithGoogle = async (): Promise<void> => {
  const { error } = await supabase.auth.signInWithOAuth({
    provider: 'google',
    options: { redirectTo: 'http://localhost:5173/auth/callback' },
  });
  if (error) throw new Error(error.message);
};

// ---- SYNC (Google OAuth callback) ----
export const syncGoogleUser = async (): Promise<boolean> => {
  const { data: { session } } = await supabase.auth.getSession();
  if (!session) return false;

  const res = await fetch(`${API_URL}/auth/sync`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${session.access_token}`
    }
  });

  return res.ok;
};