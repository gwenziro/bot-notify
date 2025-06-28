import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import apiClient from '../api/client';

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
}

export interface MessageStats {
  totalSent: number;
  todaySent: number;
  successRate: number;
  lastSentTime?: string;
}

interface DashboardState {
  // State
  connectionStatus: ConnectionStatus | null;
  qrCode: QRCodeData | null;
  messageStats: MessageStats | null;
  isRefreshing: boolean;
  lastUpdated: Date | null;
  error: string | null;

  // Actions
  setConnectionStatus: (status: ConnectionStatus) => void;
  setQRCode: (qrCode: QRCodeData) => void;
  setMessageStats: (stats: MessageStats) => void;
  setRefreshing: (refreshing: boolean) => void;
  setError: (error: string | null) => void;
  refreshAll: () => Promise<void>;
  refreshQRCode: () => Promise<void>;
  clearData: () => void;
}

export const useDashboardStore = create<DashboardState>()(
  devtools(
    (set, get) => ({
      // Initial state
      connectionStatus: null,
      qrCode: null,
      messageStats: null,
      isRefreshing: false,
      lastUpdated: null,
      error: null,

      // Actions
      setConnectionStatus: (status) =>
        set(
          { connectionStatus: status, lastUpdated: new Date(), error: null },
          false,
          'setConnectionStatus'
        ),

      setQRCode: (qrCode) =>
        set(
          { qrCode, lastUpdated: new Date(), error: null },
          false,
          'setQRCode'
        ),

      setMessageStats: (stats) =>
        set(
          { messageStats: stats, lastUpdated: new Date(), error: null },
          false,
          'setMessageStats'
        ),

      setRefreshing: (refreshing) =>
        set({ isRefreshing: refreshing }, false, 'setRefreshing'),

      setError: (error) =>
        set({ error }, false, 'setError'),

      refreshAll: async () => {
        const { setRefreshing, setConnectionStatus, setQRCode, setMessageStats, setError } = get();
        
        setRefreshing(true);
        setError(null);

        try {
          // Parallel API calls untuk performa yang lebih baik
          const [statusRes, qrRes, statsRes] = await Promise.allSettled([
            apiClient.get<ConnectionStatus>('/api/status'),
            apiClient.get<QRCodeData>('/api/qr/status'),
            apiClient.get<MessageStats>('/api/stats'),
          ]);

          // Process results
          if (statusRes.status === 'fulfilled' && statusRes.value.success) {
            setConnectionStatus(statusRes.value.data!);
          }

          if (qrRes.status === 'fulfilled' && qrRes.value.success) {
            setQRCode(qrRes.value.data!);
          }

          if (statsRes.status === 'fulfilled' && statsRes.value.success) {
            setMessageStats(statsRes.value.data!);
          }

          // Check for any failures
          const failures = [statusRes, qrRes, statsRes].filter(
            (result) => result.status === 'rejected'
          );

          if (failures.length > 0) {
            console.warn('Some API calls failed:', failures);
            setError('Beberapa data gagal dimuat');
          }
        } catch (error) {
          console.error('Failed to refresh dashboard data:', error);
          setError('Gagal memuat data dashboard');
        } finally {
          setRefreshing(false);
        }
      },

      refreshQRCode: async () => {
        const { setError } = get();
        
        try {
          setError(null);
          const response = await apiClient.post('/api/qr/refresh');
          
          if (response.success) {
            // Refresh QR status setelah refresh berhasil
            const qrResponse = await apiClient.get<QRCodeData>('/api/qr/status');
            if (qrResponse.success) {
              get().setQRCode(qrResponse.data!);
            }
          }
        } catch (error) {
          console.error('Failed to refresh QR code:', error);
          setError('Gagal menyegarkan QR code');
          throw error;
        }
      },

      clearData: () =>
        set(
          {
            connectionStatus: null,
            qrCode: null,
            messageStats: null,
            isRefreshing: false,
            lastUpdated: null,
            error: null,
          },
          false,
          'clearData'
        ),
    }),
    {
      name: 'dashboard-store',
    }
  )
);