import { RefreshCw, CheckCircle, AlertCircle, Clock } from 'lucide-react';
import { useSyncStatus } from '../../hooks/useSyncStatus';

/**
 * Format date to relative time (e.g., "2 minuty temu")
 */
const formatRelativeTime = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'przed chwilą';
  if (diffMins < 60) return `${diffMins} ${diffMins === 1 ? 'minutę' : 'minut'} temu`;
  if (diffHours < 24) return `${diffHours} ${diffHours === 1 ? 'godzinę' : 'godzin'} temu`;
  return `${diffDays} ${diffDays === 1 ? 'dzień' : 'dni'} temu`;
};

/**
 * SyncStatusIndicator component
 * Displays current sync status in the header
 */
export const SyncStatusIndicator = () => {
  const { data: syncStatus, isLoading, isError } = useSyncStatus();

  // Error state
  if (isError) {
    return (
      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <AlertCircle className="w-4 h-4 text-red-500" />
        <span>Błąd synchronizacji</span>
      </div>
    );
  }

  // Loading state
  if (isLoading || !syncStatus) {
    return (
      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <Clock className="w-4 h-4 animate-pulse" />
        <span>Sprawdzanie...</span>
      </div>
    );
  }

  // Currently syncing
  if (syncStatus.current_run) {
    return (
      <div className="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400">
        <RefreshCw className="w-4 h-4 animate-spin" />
        <span>Synchronizacja w toku...</span>
      </div>
    );
  }

  // Last sync info
  if (syncStatus.last_sync_at) {
    return (
      <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
        <CheckCircle className="w-4 h-4" />
        <span>Zsynchronizowano {formatRelativeTime(syncStatus.last_sync_at)}</span>
      </div>
    );
  }

  // No sync yet
  return (
    <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
      <Clock className="w-4 h-4" />
      <span>Brak synchronizacji</span>
    </div>
  );
};
