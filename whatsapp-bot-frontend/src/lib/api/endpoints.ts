export const API_ENDPOINTS = {
  // Authentication
  AUTH: {
    LOGIN: '/api/auth/login',
    LOGOUT: '/api/auth/logout',
    REFRESH: '/api/auth/refresh',
    VERIFY: '/api/auth/verify',
  },
  
  // Connection Management
  CONNECTION: {
    STATUS: '/api/status',
    RECONNECT: '/api/reconnect',
    DISCONNECT: '/api/disconnect',
  },
  
  // QR Code
  QR: {
    STATUS: '/api/qr/status',
    IMAGE: '/api/qr/image',
    REFRESH: '/api/qr/refresh',
  },
  
  // Messaging
  MESSAGING: {
    SEND_PERSONAL: '/api/send/personal',
    SEND_GROUP: '/api/send/group',
    BROADCAST: '/api/send/broadcast',
    HISTORY: '/api/messages/history',
  },
  
  // Groups
  GROUPS: {
    LIST: '/api/groups',
    PARTICIPANTS: (id: string) => `/api/groups/${id}/participants`,
    INFO: (id: string) => `/api/groups/${id}`,
  },
  
  // Profile
  PROFILE: {
    GET: '/api/profile',
    UPDATE: '/api/profile',
  },
  
  // Admin (Multi-user)
  ADMIN: {
    USERS_STATUS: '/api/admin/users/status',
    USER_DETAIL: (userId: string) => `/api/admin/users/${userId}`,
  },
  
  // Statistics
  STATS: '/api/stats',
} as const;