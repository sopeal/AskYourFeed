import { useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import type { CreateQACommand, QADetailDTO, ErrorResponseDTO } from '../types/qa.types';

/**
 * API client configuration
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Ask a question about the user's feed
 * POST /api/v1/qa
 */
const askQuestion = async (data: CreateQACommand): Promise<QADetailDTO> => {
  const response = await axios.post<QADetailDTO>(
    `${API_BASE_URL}/api/v1/qa`,
    data,
    {
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true, // Include session cookie
    }
  );
  return response.data;
};

/**
 * Hook for asking questions about the feed
 * Handles API call, loading state, and error handling
 */
export const useAskQuestion = () => {
  return useMutation<QADetailDTO, AxiosError<ErrorResponseDTO>, CreateQACommand>({
    mutationFn: askQuestion,
  });
};
