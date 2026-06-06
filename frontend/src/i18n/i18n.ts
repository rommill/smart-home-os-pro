/**
 * Strict structure definition for dynamic interface translation envelopes
 */
interface TranslationDictionary {
  [key: string]: string;
}

interface LanguagePayload {
  [lang: string]: TranslationDictionary;
}

// Thread-safe production dictionaries aligned with Estonia local deployment rules
const translations: LanguagePayload = {
  ru: {
    statusLoading: "Загрузка...",
    statusOnline: "Система онлайн",
    statusConnErr: "Ошибка соединения",
    statusAuthErr: "Сессия истекла",
    statusNeedAuth: "Требуется авторизация",
    statusAuthenticating: "Проверка данных...",
    statusSuccessAuth: "Вход успешно выполнен",
    statusLogout: "Выход из системы",
    statusRequired: "Пожалуйста, войдите в систему",
    loginTitle: "Вход в систему",
    usernameLabel: "Имя пользователя",
    passwordLabel: "Пароль",
    loginBtn: "Войти",
    logoutBtn: "Выйти",
    targetTemp: "Целевая температура",
  },
  et: {
    statusLoading: "Laadimine...",
    statusOnline: "Süsteem on võrgus",
    statusConnErr: "Ühenduse viga",
    statusAuthErr: "Sessioon on aegunud",
    statusNeedAuth: "Vajalik autoriseerimine",
    statusAuthenticating: "Andmete kontrollimine...",
    statusSuccessAuth: "Sisselogimine õnnestus",
    statusLogout: "Süsteemist välja logitud",
    statusRequired: "Palun logige sisse",
    loginTitle: "Sisselogimine",
    usernameLabel: "Kasutajatunnus",
    passwordLabel: "Parool",
    loginBtn: "Logi sisse",
    logoutBtn: "Logi välja",
    targetTemp: "Sihttemperatuur",
  },
  en: {
    statusLoading: "Loading...",
    statusOnline: "System Online",
    statusConnErr: "Connection Error",
    statusAuthErr: "Session Expired",
    statusNeedAuth: "Authorization Required",
    statusAuthenticating: "Authenticating...",
    statusSuccessAuth: "Authentication Successful",
    statusLogout: "Logged Out",
    statusRequired: "Please log in to continue",
    loginTitle: "System Authentication",
    usernameLabel: "Username",
    passwordLabel: "Password",
    loginBtn: "Sign In",
    logoutBtn: "Sign Out",
    targetTemp: "Target Temperature",
  },
};

const STORAGE_LANG_KEY = "smarthome_lang";
let currentLang: string = localStorage.getItem(STORAGE_LANG_KEY) || "en";
const listeners: (() => void)[] = [];

export function getCurrentLang(): string {
  return currentLang;
}

export function setLang(lang: string): void {
  if (translations[lang]) {
    currentLang = lang;
    localStorage.setItem(STORAGE_LANG_KEY, lang);
    // Notify all bound execution contexts about runtime localization mutations
    listeners.forEach((callback) => callback());
  }
}

export function registerLangChangeListener(callback: () => void): void {
  listeners.push(callback);
}

/**
 * Resolves localized dictionary key transformations based on current active bounds.
 */
export function t(key: string): string {
  const dictionary = translations[currentLang];
  if (dictionary && dictionary[key]) {
    return dictionary[key];
  }
  // Fallback cascade logic to prevent layout crashing on undefined labels
  return translations["en"][key] || key;
}

/**
 * Scans the active DOM tree and rewires translating labels dynamically
 */
export function translatePage(): void {
  const elements = document.querySelectorAll("[data-i18n]");
  elements.forEach((el) => {
    const htmlEl = el as HTMLElement;
    const key = htmlEl.getAttribute("data-i18n");
    if (key) {
      htmlEl.innerText = t(key);
    }
  });
}
