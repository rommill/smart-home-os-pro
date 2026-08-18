import { RoomData, UpdateResponse } from "../types/telemetry";

const API_URL = "http://127.0.0.1:8080";

interface LoginResponse {
  token: string;
}

/**
 * Handles automatic session cleanup when an expired or invalid token is rejected by the server
 */
function handleUnauthorized(): void {
  console.warn("⚠️ [Auth] Session expired or invalid (401). Clearing token...");
  localStorage.removeItem("jwt_token");
  localStorage.removeItem("token");
  window.location.reload();
}

/**
 * Formats the raw JWT string into a standard Bearer Authorization header value
 */
function getAuthHeader(token: string): string {
  return token.startsWith("Bearer ") ? token : `Bearer ${token}`;
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

  if (!response.ok) {
    throw new Error("Invalid username or password");
  }

  return (await response.json()) as LoginResponse;
}

/**
 * Fetches the latest sensor telemetry array from the backend
 */
export async function fetchTelemetry(token: string): Promise<RoomData[]> {
  const response = await fetch(`${API_URL}/telemetry`, {
    method: "GET",
    headers: {
      Authorization: getAuthHeader(token),
      "Content-Type": "application/json",
    },
  });

  if (response.status === 401) {
    handleUnauthorized();
    throw new Error("Unauthorized");
  }
  if (!response.ok) {
    throw new Error("Server error while fetching telemetry");
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
      Authorization: getAuthHeader(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      target_temperature: targetTemperature,
    }),
  });

  if (response.status === 401) {
    handleUnauthorized();
    throw new Error("Unauthorized");
  }
  if (!response.ok) {
    throw new Error("Failed to update target temperature on the server");
  }

  return (await response.json()) as UpdateResponse;
}
