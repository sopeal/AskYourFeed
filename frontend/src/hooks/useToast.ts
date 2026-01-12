import { useState, useCallback } from 'react';

export interface ToastMessage {
  id: string;
  message: string;
  variant: 'default' | 'destructive' | 'success';
}

/**
 * Hook for managing toast notifications
 * Provides methods to show and dismiss toasts
 */
export const useToast = () => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  /**
   * Show a toast notification
   * Auto-dismisses after 5 seconds
   */
  const showToast = useCallback((message: string, variant: ToastMessage['variant'] = 'default') => {
    const id = Math.random().toString(36).substring(7);
    const toast: ToastMessage = { id, message, variant };
    
    setToasts((prev) => [...prev, toast]);

    // Auto-dismiss after 5 seconds
    setTimeout(() => {
      dismissToast(id);
    }, 5000);

    return id;
  }, []);

  /**
   * Dismiss a specific toast
   */
  const dismissToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  /**
   * Show error toast
   */
  const showError = useCallback((message: string) => {
    return showToast(message, 'destructive');
  }, [showToast]);

  /**
   * Show success toast
   */
  const showSuccess = useCallback((message: string) => {
    return showToast(message, 'success');
  }, [showToast]);

  return {
    toasts,
    showToast,
    showError,
    showSuccess,
    dismissToast,
  };
};
