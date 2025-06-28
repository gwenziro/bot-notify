import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConnectionStatus } from '../connection-status';
import apiClient from '@/lib/api/client';

// Mock API client
jest.mock('@/lib/api/client');
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>;

// Mock hooks
jest.mock('@/lib/hooks/use-notification', () => ({
  useNotification: () => ({
    showNotification: jest.fn(),
  }),
}));

jest.mock('@/lib/hooks/use-websocket', () => ({
  useWebSocket: () => ({
    subscribe: jest.fn(() => jest.fn()),
  }),
}));

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

const renderWithQueryClient = (component: React.ReactElement) => {
  const testQueryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={testQueryClient}>
      {component}
    </QueryClientProvider>
  );
};

describe('ConnectionStatus', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockedApiClient.get.mockImplementation(() => new Promise(() => {})); // Never resolves
    
    renderWithQueryClient(<ConnectionStatus />);
    
    expect(screen.getByText('Memuat Status Koneksi...')).toBeInTheDocument();
  });

  it('renders connected state correctly', async () => {
    const mockStatus = {
      success: true,
      data: {
        isConnected: true,
        status: 'connected',
        lastActivity: '2023-12-01T10:00:00Z',
        connectedSince: '2023-12-01T09:00:00Z',
        messagesSent: 42,
      },
    };

    mockedApiClient.get.mockResolvedValue(mockStatus);

    renderWithQueryClient(<ConnectionStatus />);

    await waitFor(() => {
      expect(screen.getByText('Status Koneksi WhatsApp')).toBeInTheDocument();
      expect(screen.getByText('Terhubung')).toBeInTheDocument();
      expect(screen.getByText('connected')).toBeInTheDocument();
      expect(screen.getByText('42')).toBeInTheDocument();
    });
  });

  it('renders disconnected state correctly', async () => {
    const mockStatus = {
      success: true,
      data: {
        isConnected: false,
        status: 'disconnected',
        lastActivity: '2023-12-01T10:00:00Z',
      },
    };

    mockedApiClient.get.mockResolvedValue(mockStatus);

    renderWithQueryClient(<ConnectionStatus />);

    await waitFor(() => {
      expect(screen.getByText('Terputus')).toBeInTheDocument();
      expect(screen.getByText('Hubungkan Ulang')).toBeInTheDocument();
    });
  });

  it('handles reconnect action', async () => {
    const mockStatus = {
      success: true,
      data: {
        isConnected: false,
        status: 'disconnected',
        lastActivity: '2023-12-01T10:00:00Z',
      },
    };

    mockedApiClient.get.mockResolvedValue(mockStatus);
    mockedApiClient.post.mockResolvedValue({ success: true });

    renderWithQueryClient(<ConnectionStatus />);

    await waitFor(() => {
      expect(screen.getByText('Hubungkan Ulang')).toBeInTheDocument();
    });

    const reconnectButton = screen.getByText('Hubungkan Ulang');
    fireEvent.click(reconnectButton);

    await waitFor(() => {
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/reconnect');
    });
  });

  it('handles disconnect action', async () => {
    const mockStatus = {
      success: true,
      data: {
        isConnected: true,
        status: 'connected',
        lastActivity: '2023-12-01T10:00:00Z',
      },
    };

    mockedApiClient.get.mockResolvedValue(mockStatus);
    mockedApiClient.post.mockResolvedValue({ success: true });

    renderWithQueryClient(<ConnectionStatus />);

    await waitFor(() => {
      expect(screen.getByText('Putuskan Koneksi')).toBeInTheDocument();
    });

    const disconnectButton = screen.getByText('Putuskan Koneksi');
    fireEvent.click(disconnectButton);

    await waitFor(() => {
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/disconnect');
    });
  });

  it('handles API errors gracefully', async () => {
    mockedApiClient.get.mockRejectedValue(new Error('Network error'));

    renderWithQueryClient(<ConnectionStatus />);

    // Component should still render even with API error
    await waitFor(() => {
      expect(screen.getByText('Status Koneksi WhatsApp')).toBeInTheDocument();
    });
  });
});