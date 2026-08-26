import { fetchTelemetry, updateTargetTemperatureRequest } from "../api/api";
import { getToken, removeToken } from "../auth/auth";
import { updateStatus, renderRooms, toggleAuthView } from "../render/render";
import { RoomData } from "../types/telemetry";

let telemetryInterval: number | null = null;

export function stopPolling(): void {
  if (telemetryInterval !== null) {
    clearInterval(telemetryInterval);
    telemetryInterval = null;
  }
}

export function startPolling(): void {
  stopPolling();
  telemetryInterval = window.setInterval(checkAndRefreshTelemetry, 3000);
}

export function handleSessionExpiry(): void {
  removeToken();
  stopPolling();
  toggleAuthView(true);
  updateStatus("statusAuthErr", "error");
}

export async function checkAndRefreshTelemetry(): Promise<void> {
  const token: string | null = getToken();
  if (!token) {
    handleSessionExpiry();
    return;
  }

  try {
    const data: RoomData[] = await fetchTelemetry(token);

    // console.log("RAW TELEMETRY RESPONSE:", data);

    if (Array.isArray(data)) {
      renderRooms(data);
      updateStatus("statusOnline", "success");
      toggleAuthView(false);
    }
  } catch (err: any) {
    if (err.message === "Unauthorized") {
      handleSessionExpiry();
    } else {
      updateStatus("statusConnErr", "error");
    }
  }
}

export async function syncTargetTemperature(
  roomId: number,
  value: string,
): Promise<void> {
  const token: string | null = getToken();
  if (!token) return;

  try {
    const numericValue: number = parseFloat(value);
    await updateTargetTemperatureRequest(roomId, numericValue, token);
  } catch (err: any) {
    if (err.message === "Unauthorized") {
      handleSessionExpiry();
    }
  }
}
