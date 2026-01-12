import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import type { IngestStatusDTO } from '../types/ingest.types';

/**
 * API client configuration
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Fetch current sync/ingest status
 * GET /api/v1/ingest/status
 */
const fetchSyncStatus = async (): Promise<IngestStatusDTO> => {
  const sessionToken = localStorage.getItem('session_token');
  
  const response = await axios.get<IngestStatusDTO>(
    `${API_BASE_URL}/api/v1/ingest/status`,
    {
      headers: {
        'Content-Type': 'application/json',
        ...(sessionToken && { 'Authorization': `Bearer ${sessionToken}` }),
      },
      withCredentials: true, // Include session cookie
    }
  );
  return response.data;
};

/**
 * Hook for fetching sync status with automatic polling
 * Polls every 30 seconds to keep status up-to-date
 * Errors are handled silently to avoid disrupting the UI
 */
export const useSyncStatus = () => {
  return useQuery<IngestStatusDTO, Error>({
    queryKey: ['syncStatus'],
    queryFn: fetchSyncStatus,
    refetchInterval: 30000, // Refetch every 30 seconds
    retry: 1, // Only retry once on failure
    staleTime: 25000, // Consider data stale after 25 seconds
  });
};
