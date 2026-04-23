export const API_CONFIG = {
  AUTH: import.meta.env.VITE_AUTH_API_URL || 'http://localhost:8084',
  BOOKING: import.meta.env.VITE_BOOKING_API_URL || 'http://localhost:8081',
  CORE: import.meta.env.VITE_CORE_API_URL || 'http://localhost:8080',
};

export const ROLES = {
  CLIENT: 'client',
  PERFORMER: 'performer',
  ADMIN: 'admin',
};