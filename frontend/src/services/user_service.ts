import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";
import axios , {InternalAxiosRequestConfig} from 'axios';

const api = axios.create({
    baseURL: "/api/users",
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

export async function getRemoteUserStats(): Promise<UserStats> {
        // const response = await api.get("/me/stats");
        // return response.data;
        return new Promise((resolve) => {
        setTimeout(() => {
            resolve({
                id: 1, // Just a mock implementation, using token length as id
                name: "pepe",
                energy: 100,
                max_energy: 500,
                level: 5
            });
        }, 500);
        });
    }

export async function getUserOrders(): Promise<UserOrder[]> {
        const response = await api.get("/me/orders/active");
        return response.data;
    }

export async function completeOrder(orderId: number): Promise<UserStats> {
        const response = await api.post(`/orders/${orderId}/complete`);
        return response.data;
    }