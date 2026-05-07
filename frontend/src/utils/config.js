export const API_CONFIG = {
  AUTH: import.meta.env.VITE_AUTH_API_URL || 'http://localhost:81',
  BOOKING: import.meta.env.VITE_BOOKING_API_URL || 'http://localhost:81',
  CORE: import.meta.env.VITE_CORE_API_URL || 'http://localhost:81',
};

export const ROLES = {
  CLIENT: 'client',
  PERFORMER: 'performer',
  ADMIN: 'admin',
};