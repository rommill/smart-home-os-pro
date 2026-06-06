const API_URL = "http://localhost:8080";
/**
 * Sends an authentication request to the Go identity provider
 */
export async function loginRequest(username, password) {
    const response = await fetch(`${API_URL}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
    });
    if (!response.ok)
        throw new Error("Неверный логин или пароль");
    return (await response.json());
}
/**
 * Fetches the latest sensor telemetry array from the backend
 */
export async function fetchTelemetry(token) {
    const response = await fetch(`${API_URL}/telemetry`, {
        method: "GET",
        headers: {
            Authorization: token,
            "Content-Type": "application/json",
        },
    });
    if (response.status === 401) {
        throw new Error("Unauthorized");
    }
    if (!response.ok) {
        throw new Error("Ошибка сервера");
    }
    return (await response.json());
}
/**
 * Syncs the updated target climate temperature back to the PostgreSQL instance
 */
export async function updateTargetTemperatureRequest(roomId, targetTemperature, token) {
    const response = await fetch(`${API_URL}/rooms/${roomId}/target-temp`, {
        method: "POST",
        headers: {
            Authorization: token,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            target_temperature: targetTemperature,
        }),
    });
    if (response.status === 401) {
        throw new Error("Unauthorized");
    }
    if (!response.ok) {
        throw new Error("Failed to update target temperature on the server");
    }
    return (await response.json());
}
