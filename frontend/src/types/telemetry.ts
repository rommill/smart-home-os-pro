/**
 * Localized name mapping enforced by the backend translation layers
 */
export interface LocalizedName {
  en: string;
  ru: string;
  et?: string;
  [key: string]: string | undefined;
}

/**
 * Unified data structure representing a smart home room instance.
 * Mirrors the Go backend domains and PostgreSQL / MongoDB combined payload.
 */
export interface RoomData {
  id: number;
  name: LocalizedName;
  temperature: string;
  last_update: string;
  target_temperature: number;
  ac_status: boolean;
}

/**
 * Standard backend response envelope for mutating operations
 */
export interface UpdateResponse {
  message: string;
  room_id: number;
  target_temperature: number;
}
