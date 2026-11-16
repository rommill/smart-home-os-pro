import { fetchTelemetry, updateTargetTemperatureRequest } from "../api/api";
import { getToken, removeToken } from "../auth/auth";
import { updateStatus, renderRooms, toggleAuthView } from "../render/render";
import { RoomData } from "../types/telemetry";

let telemetryInterval: number | null = null;
let ws: WebSocket | null = null;

// RU: Остановка поллинга по REST API
// EN: Stop REST API polling
export function stopPolling(): void {
  if (telemetryInterval !== null) {
    clearInterval(telemetryInterval);
    telemetryInterval = null;
  }
}

// RU: Запуск поллинга по REST API (резервный механизм)
// EN: Start REST API polling (fallback mechanism)
export function startPolling(): void {
  stopPolling();
  telemetryInterval = window.setInterval(checkAndRefreshTelemetry, 5000);
}

// RU: Обработка истечения авторизационной сессии
// EN: Handle expired authentication session
export function handleSessionExpiry(): void {
  removeToken();
  stopPolling();
  closeWebSocket();
  toggleAuthView(true);
  updateStatus("statusAuthErr", "error");
}

// RU: Закрытие активного WebSocket соединения
// EN: Close active WebSocket connection
export function closeWebSocket(): void {
  if (ws) {
    ws.close();
    ws = null;
  }
}

// RU: Инициализация и обработка WebSocket соединения для живых метрик
// EN: Initialize and handle WebSocket connection for live metrics
export function initWebSocket(): void {
  const token = getToken();
  if (!token) {
    handleSessionExpiry();
    return;
  }

  closeWebSocket();

  // RU: Формируем адрес WS сервера на основе текущего хоста
  // EN: Build WS server address based on current window location
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.hostname}:8080/ws`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    // console.log("⚡ [WS] Connected to live telemetry stream");
    updateStatus("statusOnline", "success");
    toggleAuthView(false);
    stopPolling(); // RU: Отключаем поллинг, так как работает WebSocket / EN: Stop polling as WS is active
  };

  ws.onmessage = (event) => {
    try {
      const liveData = JSON.parse(event.data);
      // console.log("📡 [WS Received]:", liveData);

      // RU: Перезапрашиваем состояния комнат или обновляем UI на основе данных
      // EN: Refresh room states or update UI based on incoming telemetry
      checkAndRefreshTelemetry();
    } catch (err) {
      // console.error("❌ [WS Error] Parsing payload failed:", err);
    }
  };

  ws.onerror = () => {
    updateStatus("statusConnErr", "error");
  };

  ws.onclose = () => {
    // console.warn("⚠️ [WS] Connection closed. Falling back to polling...");
    startPolling(); // RU: При разрыве соединения возвращаемся на поллинг / EN: Fallback to polling on disconnect
  };
}

// RU: Проверка и обновление телеметрии по REST
// EN: Check and refresh telemetry via REST
export async function checkAndRefreshTelemetry(): Promise<void> {
  const token: string | null = getToken();
  if (!token) {
    handleSessionExpiry();
    return;
  }

  try {
    const data: RoomData[] = await fetchTelemetry(token);

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

// RU: Синхронизация целевой температуры с бэкендом
// EN: Sync target temperature with backend
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
