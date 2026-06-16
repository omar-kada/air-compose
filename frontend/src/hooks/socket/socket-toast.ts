// lib/ws-toast.ts
import type { Namespace, TFunction } from 'i18next';
import { toast } from 'sonner';
import type { WsStatus } from './socket-provider';

let toastId: string | number | null = 'WS_TOAST';

const PERSISTENT = {
  duration: Infinity,
  dismissible: false,
} as const;

export const wsToast = {
  connected: (t: TFunction) => {
    toastId = toast.success(t('SOCKET.CONNECTED.TITLE'), {
      id: toastId ?? undefined,
      duration: 4000,
      description: t('SOCKET.CONNECTED.DESCRIPTION'),
    });
  },

  reconnecting: (t: TFunction, attempt: number, max: number) => {
    toastId = toast.loading(t('SOCKET.RECONNECTING.TITLE'), {
      ...PERSISTENT,
      id: toastId ?? undefined,
      description: t('SOCKET.RECONNECTING.DESCRIPTION', { attempt, max }),
    });
  },

  failed: (t: TFunction) => {
    toastId = toast.error(t('SOCKET.FAILED.TITLE'), {
      ...PERSISTENT,
      id: toastId ?? undefined,
      description: t('SOCKET.FAILED.DESCRIPTION'),
      action: {
        label: t('SOCKET.FAILED.ACTION'),
        onClick: () => window.location.reload(),
      },
    });
  },

  dismiss: () => {
    if (toastId !== null) {
      toast.dismiss(toastId);
      toastId = null;
    }
  },
};

export function wsToastOnStatus(
  t: TFunction<Namespace>,
  oldStatus: WsStatus,
  status: WsStatus,
  attempt?: number,
  max?: number,
) {
  switch (status) {
    case 'off':
      wsToast.dismiss();
      break;
    case 'connected':
      if (oldStatus === 'reconnecting') {
        wsToast.connected(t);
      } else {
        wsToast.dismiss();
      }
      break;
    case 'reconnecting':
      wsToast.reconnecting(t, attempt ?? 0, max ?? 0);
      break;
    case 'failed':
      wsToast.failed(t);
      break;
  }
}
