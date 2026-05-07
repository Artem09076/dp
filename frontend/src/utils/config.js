export const API_CONFIG = {
  AUTH: import.meta.env.VITE_AUTH_API_URL || '',
  BOOKING: import.meta.env.VITE_BOOKING_API_URL || '',
  CORE: import.meta.env.VITE_CORE_API_URL || '',
};

export const ROLES = {
  CLIENT: 'client',
  PERFORMER: 'performer',
  ADMIN: 'admin',
};