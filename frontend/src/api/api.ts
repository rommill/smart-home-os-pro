import { RoomData, UpdateResponse } from "../types/telemetry.js";

const API_URL = "http://localhost:8080";

interface LoginResponse {
  token: string;
}

/**
 * Sends an authentication request to the Go identity provider
 */
export async function loginRequest(
  username: string,
  password: string,
): Promise<LoginResponse> {
  const response = await fetch(`${API_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) throw new Error("Неверный логин или пароль");
  return (await response.json()) as LoginResponse;
}

/**
 * Fetches the latest sensor telemetry array from the backend
 */
export async function fetchTelemetry(token: string): Promise<RoomData[]> {
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

  return (await response.json()) as RoomData[];
}

/**
 * Syncs the updated target climate temperature back to the PostgreSQL instance
 */
export async function updateTargetTemperatureRequest(
  roomId: number,
  targetTemperature: number,
  token: string,
): Promise<UpdateResponse> {
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

  return (await response.json()) as UpdateResponse;
}
