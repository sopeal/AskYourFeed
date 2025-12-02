import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import type { 
  QAListResponseDTO, 
  QADetailDTO, 
  ErrorResponseDTO 
} from '../../../types/qa.types';

/**
 * API client configuration
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Fetch paginated Q&A history list
 * GET /api/v1/qa
 */
const fetchQAHistory = async ({ pageParam }: { pageParam?: string }): Promise<QAListResponseDTO> => {
  const params = new URLSearchParams();
  params.append('limit', '20');
  if (pageParam) {
    params.append('cursor', pageParam);
  }

  const response = await axios.get<QAListResponseDTO>(
    `${API_BASE_URL}/api/v1/qa?${params.toString()}`,
    {
      withCredentials: true,
    }
  );
  return response.data;
};

/**
 * Fetch single Q&A detail
 * GET /api/v1/qa/{id}
 */
const fetchQADetail = async (id: string): Promise<QADetailDTO> => {
  const response = await axios.get<QADetailDTO>(
    `${API_BASE_URL}/api/v1/qa/${id}`,
    {
      withCredentials: true,
    }
  );
  return response.data;
};

/**
 * Delete single Q&A entry
 * DELETE /api/v1/qa/{id}
 */
const deleteQAEntry = async (id: string): Promise<{ message: string }> => {
  const response = await axios.delete(
    `${API_BASE_URL}/api/v1/qa/${id}`,
    {
      withCredentials: true,
    }
  );
  return response.data;
};

/**
 * Delete all Q&A history
 * DELETE /api/v1/qa
 */
const deleteAllQAHistory = async (): Promise<{ message: string; deleted_count: number }> => {
  const response = await axios.delete(
    `${API_BASE_URL}/api/v1/qa`,
    {
      withCredentials: true,
    }
  );
  return response.data;
};

/**
 * Hook for fetching Q&A detail by ID
 * Used for lazy loading full answer when accordion is expanded
 */
export const useQADetail = (id: string, enabled: boolean = false) => {
  return useQuery<QADetailDTO, AxiosError<ErrorResponseDTO>>({
    queryKey: ['qa-detail', id],
    queryFn: () => fetchQADetail(id),
    enabled,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

/**
 * Main hook for Q&A history management
 * Provides infinite scroll pagination, deletion mutations, and state management
 */
export const useHistory = () => {
  const queryClient = useQueryClient();

  // Infinite query for paginated history list
  const historyQuery = useInfiniteQuery<QAListResponseDTO, AxiosError<ErrorResponseDTO>>({
    queryKey: ['qa-history'],
    queryFn: fetchQAHistory,
    getNextPageParam: (lastPage) => {
      return lastPage.has_more ? lastPage.next_cursor : undefined;
    },
    initialPageParam: undefined,
  });

  // Mutation for deleting single entry
  const deleteOneMutation = useMutation<
    { message: string },
    AxiosError<ErrorResponseDTO>,
    string
  >({
    mutationFn: deleteQAEntry,
    onSuccess: () => {
      // Invalidate and refetch history list
      queryClient.invalidateQueries({ queryKey: ['qa-history'] });
    },
  });

  // Mutation for deleting all history
  const deleteAllMutation = useMutation<
    { message: string; deleted_count: number },
    AxiosError<ErrorResponseDTO>
  >({
    mutationFn: deleteAllQAHistory,
    onSuccess: () => {
      // Invalidate and refetch history list
      queryClient.invalidateQueries({ queryKey: ['qa-history'] });
    },
  });

  return {
    // History list data and state
    data: historyQuery.data,
    isLoading: historyQuery.isLoading,
    isError: historyQuery.isError,
    error: historyQuery.error,
    
    // Pagination
    hasNextPage: historyQuery.hasNextPage,
    isFetchingNextPage: historyQuery.isFetchingNextPage,
    fetchNextPage: historyQuery.fetchNextPage,
    
    // Delete mutations
    deleteOne: deleteOneMutation.mutate,
    deleteAll: deleteAllMutation.mutate,
    isDeletingOne: deleteOneMutation.isPending,
    isDeletingAll: deleteAllMutation.isPending,
    deleteError: deleteOneMutation.error || deleteAllMutation.error,
  };
};
