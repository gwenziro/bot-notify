'use client';

import { useConnectionStatus } from '@/lib/hooks/use-connection-status';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loader2, Wifi, WifiOff, RotateCcw, Power } from 'lucide-react';
import { useState } from 'react';
import { useNotification } from '@/lib/hooks/use-notification';

export function ConnectionStatus() {
  const { status, isLoading, reconnect, disconnect } = useConnectionStatus();
  const [isReconnecting, setIsReconnecting] = useState(false);
  const [isDisconnecting, setIsDisconnecting] = useState(false);
  const { showNotification } = useNotification();

  const handleReconnect = async () => {
    setIsReconnecting(true);
    try {
      await reconnect();
      showNotification('Berhasil memulai proses reconnect', 'success');
    } catch (error) {
      showNotification('Gagal melakukan reconnect', 'error');
    } finally {
      setIsReconnecting(false);
    }
  };

  const handleDisconnect = async () => {
    setIsDisconnecting(true);
    try {
      await disconnect();
      showNotification('WhatsApp berhasil diputuskan', 'success');
    } catch (error) {
      showNotification('Gagal memutuskan koneksi', 'error');
    } finally {
      setIsDisconnecting(false);
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Loader2 className="h-5 w-5 animate-spin" />
            Memuat Status Koneksi...
          </CardTitle>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {status?.isConnected ? (
              <Wifi className="h-5 w-5 text-green-500" />
            ) : (
              <WifiOff className="h-5 w-5 text-red-500" />
            )}
            Status Koneksi WhatsApp
          </div>
          <Badge variant={status?.isConnected ? 'success' : 'destructive'}>
            {status?.isConnected ? 'Terhubung' : 'Terputus'}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="font-medium">Status:</span>
            <p className="text-muted-foreground">{status?.status || 'Unknown'}</p>
          </div>
          <div>
            <span className="font-medium">Aktivitas Terakhir:</span>
            <p className="text-muted-foreground">
              {status?.lastActivity ? new Date(status.lastActivity).toLocaleString('id-ID') : '-'}
            </p>
          </div>
          {status?.connectedSince && (
            <div>
              <span className="font-medium">Terhubung Sejak:</span>
              <p className="text-muted-foreground">
                {new Date(status.connectedSince).toLocaleString('id-ID')}
              </p>
            </div>
          )}
          {status?.messagesSent !== undefined && (
            <div>
              <span className="font-medium">Pesan Terkirim:</span>
              <p className="text-muted-foreground">{status.messagesSent}</p>
            </div>
          )}
        </div>

        <div className="flex gap-2">
          {status?.isConnected ? (
            <Button
              variant="destructive"
              size="sm"
              onClick={handleDisconnect}
              disabled={isDisconnecting}
            >
              {isDisconnecting ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <Power className="h-4 w-4 mr-2" />
              )}
              Putuskan Koneksi
            </Button>
          ) : (
            <Button
              variant="default"
              size="sm"
              onClick={handleReconnect}
              disabled={isReconnecting}
            >
              {isReconnecting ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <RotateCcw className="h-4 w-4 mr-2" />
              )}
              Hubungkan Ulang
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}