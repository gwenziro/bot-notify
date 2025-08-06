'use client';

import { createContext, useContext, useEffect, useState, useCallback, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { TokenManager } from './token-manager';
import apiClient from '../api/client';

export interface User {
  id: string;
  email?: string;
  name?: string;
  role?: string;
  isAdmin?: boolean;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (token: string, userData?: User) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  const login = useCallback(async (token: string, userData?: User) => {
    TokenManager.setTokens(token);
    
    if (userData) {
      setUser(userData);
      TokenManager.setUserData(userData);
    } else {
      // Fetch user data if not provided
      try {
        const response = await apiClient.get('/api/profile');
        if (response.success && response.data) {
          setUser(response.data);
          TokenManager.setUserData(response.data);
        }
      } catch (error) {
        console.error('Failed to fetch user data:', error);
      }
    }
  }, []);

  const logout = useCallback(() => {
    TokenManager.clearTokens();
    setUser(null);
    router.push('/login');
  }, [router]);

  const refreshAuth = useCallback(async () => {
    const token = TokenManager.getAccessToken();
    if (!token) {
      setIsLoading(false);
      return;
    }

    try {
      const response = await apiClient.get('/api/auth/verify');
      if (response.success && response.data) {
        setUser(response.data.user);
      } else {
        // Try to refresh token
        const newToken = await TokenManager.refreshAccessToken();
        if (newToken) {
          const retryResponse = await apiClient.get('/api/auth/verify');
          if (retryResponse.success && retryResponse.data) {
            setUser(retryResponse.data.user);
          } else {
            logout();
          }
        } else {
          logout();
        }
      }
    } catch (error) {
      console.error('Auth verification failed:', error);
      logout();
    } finally {
      setIsLoading(false);
    }
  }, [logout]);

  // Initialize auth state
  useEffect(() => {
    const initAuth = async () => {
      const token = TokenManager.getAccessToken();
      const userData = TokenManager.getUserData();

      if (token && userData) {
        setUser(userData);
        // Verify token is still valid
        await refreshAuth();
      } else {
        setIsLoading(false);
      }
    };

    initAuth();
  }, [refreshAuth]);

  const value = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    logout,
    refreshAuth,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}