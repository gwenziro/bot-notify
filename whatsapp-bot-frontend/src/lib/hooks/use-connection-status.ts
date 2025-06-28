import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import apiClient from '../api/client';
import { useWebSocket } from './use-websocket';

export interface ConnectionStatus {
  isConnected: boolean;
  status: string;
  lastActivity: string;
  connectedSince?: string;
  messagesSent?: number;
}

export function useConnectionStatus() {
  const queryClient = useQueryClient();
  
  // WebSocket untuk real-time updates
  const { subscribe } = useWebSocket('/ws');

  // Query untuk fetch status
  const {
    data: status,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['connection-status'],
    queryFn: async (): Promise<ConnectionStatus> => {
      const response = await apiClient.get<ConnectionStatus>('/api/status');
      return response.data || {
        isConnected: false,
        status: 'disconnected',
        lastActivity: '',
      };
    },
    refetchInterval: 30000, // Fallback polling setiap 30 detik
    staleTime: 5000, // Data dianggap fresh selama 5 detik
  });

  // Subscribe ke WebSocket updates
  useEffect(() => {
    const unsubscribe = subscribe('connection_status_changed', (newStatus: ConnectionStatus) => {
      queryClient.setQueryData(['connection-status'], newStatus);
    });

    return unsubscribe;
  }, [subscribe, queryClient]);

  // Actions
  const reconnect = async () => {
    try {
      await apiClient.post('/api/reconnect');
      // Refetch status setelah reconnect
      setTimeout(() => refetch(), 1000);
    } catch (error) {
      throw error;
    }
  };

  const disconnect = async () => {
    try {
      await apiClient.post('/api/disconnect');
      // Update status langsung tanpa menunggu
      queryClient.setQueryData(['connection-status'], {
        isConnected: false,
        status: 'disconnected',
        lastActivity: new Date().toISOString(),
      });
    } catch (error) {
      throw error;
    }
  };

  return {
    status,
    isLoading,
    error,
    refetch,
    reconnect,
    disconnect,
  };
}