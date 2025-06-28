import { useEffect, useRef, useState, useCallback } from 'react';

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

interface WebSocketMessage {
  event: string;
  data: any;
  timestamp: string;
}

export function useWebSocket(url: string) {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const listenersRef = useRef<Map<string, Function[]>>(new Map());
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;

  const connect = useCallback(() => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    setConnectionStatus('connecting');
    
    try {
      const wsUrl = url.startsWith('ws') ? url : `ws://localhost:8080${url}`;
      socketRef.current = new WebSocket(wsUrl);

      socketRef.current.onopen = () => {
        setConnectionStatus('connected');
        reconnectAttempts.current = 0;
        console.log('WebSocket connected');
      };

      socketRef.current.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          setLastMessage(message);

          // Notify listeners
          const listeners = listenersRef.current.get(message.event) || [];
          listeners.forEach((listener) => {
            try {
              listener(message.data);
            } catch (error) {
              console.error('Error in WebSocket listener:', error);
            }
          });
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      socketRef.current.onclose = (event) => {
        setConnectionStatus('disconnected');
        console.log('WebSocket disconnected:', event.code, event.reason);

        // Auto-reconnect dengan exponential backoff
        if (reconnectAttempts.current < maxReconnectAttempts) {
          const delay = Math.pow(2, reconnectAttempts.current) * 1000; // 1s, 2s, 4s, 8s, 16s
          reconnectAttempts.current++;
          
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log(`Attempting to reconnect (${reconnectAttempts.current}/${maxReconnectAttempts})`);
            connect();
          }, delay);
        }
      };

      socketRef.current.onerror = (error) => {
        setConnectionStatus('error');
        console.error('WebSocket error:', error);
      };
    } catch (error) {
      setConnectionStatus('error');
      console.error('Failed to create WebSocket connection:', error);
    }
  }, [url]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
    }
    
    setConnectionStatus('disconnected');
    reconnectAttempts.current = 0;
  }, []);

  const subscribe = useCallback((event: string, callback: Function) => {
    const listeners = listenersRef.current.get(event) || [];
    listeners.push(callback);
    listenersRef.current.set(event, listeners);

    // Return unsubscribe function
    return () => {
      const updatedListeners = listenersRef.current.get(event)?.filter(cb => cb !== callback) || [];
      if (updatedListeners.length === 0) {
        listenersRef.current.delete(event);
      } else {
        listenersRef.current.set(event, updatedListeners);
      }
    };
  }, []);

  const emit = useCallback((event: string, data: any) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = {
        event,
        data,
        timestamp: new Date().toISOString(),
      };
      socketRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Cannot send message:', { event, data });
    }
  }, []);

  // Auto-connect on mount
  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, []);

  return {
    connectionStatus,
    lastMessage,
    connect,
    disconnect,
    subscribe,
    emit,
    isConnected: connectionStatus === 'connected',
  };
}