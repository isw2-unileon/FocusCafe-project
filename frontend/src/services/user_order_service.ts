import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";
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

export async function getUserOrders(): Promise<UserOrder[]> {
    const response = await api.get("/me/orders");
    return response.data;
}

export async function completeOrder(orderId: number): Promise<UserStats> {
    const response = await api.post(`me/orders/${orderId}/complete`);
    return response.data;
}
