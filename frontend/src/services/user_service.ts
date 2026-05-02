import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";
import { UserProfile } from "@/types/user-profile";
import axios , {InternalAxiosRequestConfig} from 'axios';

const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL+"/users",
    headers: {
        "Content-Type": "application/json"
    }
})

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export async function getRemoteUserStats(): Promise<UserStats> {
        const response = await api.get("/me");
        return response.data;
    }

export async function getUserOrders(): Promise<UserOrder[]> {
        const response = await api.get("/me/orders/active");
        return response.data;
    }

export async function completeOrder(orderId: number): Promise<UserStats> {
        const response = await api.post(`/orders/${orderId}/complete`);
        return response.data;
    }

export async function getCurrentProfile(): Promise<UserProfile> {
        const response = await api.get('/me');
        return response.data;
    }

export async function updateUserProfile(data: { first_name: string; last_name: string }): Promise<UserProfile> {
        const response = await api.put('/me', data);
        return response.data;
    }