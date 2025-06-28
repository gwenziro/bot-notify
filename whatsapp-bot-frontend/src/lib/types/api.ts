export interface APIResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
  timestamp: string;
}

export interface PaginatedResponse<T> extends APIResponse<T[]> {
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ConnectionStatus {
  isConnected: boolean;
  status: string;
  lastActivity: string;
  connectedSince?: string;
  messagesSent?: number;
}

export interface QRCodeData {
  available: boolean;
  expired: boolean;
  url?: string;
  message: string;
  timestamp?: string;
}

export interface MessageStats {
  totalSent: number;
  todaySent: number;
  successRate: number;
  lastSentTime?: string;
}

export interface GroupInfo {
  id: string;
  name: string;
  memberCount: number;
  isAdmin: boolean;
  participants?: GroupParticipant[];
}

export interface GroupParticipant {
  jid: string;
  phoneNumber?: string;
  isAdmin: boolean;
  isSuperAdmin?: boolean;
  displayName?: string;
  pushName?: string;
  contactName?: string;
}

export interface SendMessageRequest {
  phoneNumber?: string;
  groupId?: string;
  message: string;
}

export interface BroadcastRequest {
  personalNumbers: string[];
  groupIds: string[];
  message: string;
  delayMs?: number;
}

export interface BroadcastResult {
  target: string;
  type: 'personal' | 'group';
  success: boolean;
  error?: string;
  sentTime?: string;
}

export interface UserProfile {
  id: string;
  phoneNumber?: string;
  name?: string;
  status?: string;
  isConnected: boolean;
  isLoggedIn: boolean;
  pictureUrl?: string;
  connectedSince?: string;
}