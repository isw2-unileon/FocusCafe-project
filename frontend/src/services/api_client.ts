const BASE_URL = import.meta.env.VITE_API_URL;

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem('token');
    
    const headers: Record<string, string> = {
        ...(token ? { "Authorization": `Bearer ${token}` } : {}),
    };

    if (!(options.body instanceof FormData)) {
        headers["Content-Type"] = "application/json";
    }

    const finalHeaders = {
        ...headers,
        ...options.headers as Record<string, string>,
    };

    const response = await fetch(`${BASE_URL}${endpoint}`, {
        ...options,
        headers: finalHeaders,
    });

    if (response.status === 401) {
        throw new Error("Session expired");
    }

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `Error: ${response.status}`);
    }

    return response.json();
}