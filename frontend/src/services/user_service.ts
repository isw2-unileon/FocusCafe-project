import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";
import { UserProfile } from "@/types/user-profile";
import { apiFetch } from "@/services/api_client";

//Prefix for all the routes
const PATH = "/users";

//Fetch remote user statistics
export async function getRemoteUserStats(): Promise<UserStats> {
    return apiFetch(`${PATH}/me`);
}

//Retrieve the list of user orders
export async function getUserOrders(): Promise<UserOrder[]> {
    return apiFetch(`${PATH}/me/orders`);
}

//Mark a specific order as complete
export async function completeOrder(orderId: number): Promise<UserStats> {
    return apiFetch(`/orders/${orderId}/complete`, {
        method: 'POST',
    });
}

//Get the current user's profile details
export async function getCurrentProfile(): Promise<UserProfile> {
    return apiFetch(`${PATH}/me`);
}

//Update the user's profile information
export async function updateUserProfile(data: { first_name: string; last_name: string }): Promise<UserProfile> {
    return apiFetch('/me', {
        method: 'PUT',
        body: JSON.stringify(data),
    });
}