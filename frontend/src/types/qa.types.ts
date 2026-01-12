/**
 * Q&A Types
 * Type definitions for Q&A functionality matching backend DTOs
 */

/**
 * Request to create a new Q&A interaction
 * POST /api/v1/qa
 */
export interface CreateQACommand {
  question: string;
  date_from?: string; // ISO 8601 format
  date_to?: string;   // ISO 8601 format
}

/**
 * Source post for Q&A answer
 * Maps to backend QASourceDTO
 */
export interface QASourceDTO {
  x_post_id: number;
  author_handle: string;
  author_display_name: string;
  published_at: string; // ISO 8601 format
  url: string;
  text_preview?: string;
  text?: string;
}

/**
 * Full Q&A interaction details
 * Maps to backend QADetailDTO
 */
export interface QADetailDTO {
  id: string;
  question: string;
  answer: string;
  date_from: string;
  date_to: string;
  created_at: string;
  sources: QASourceDTO[];
}

/**
 * Q&A item in paginated list
 * Maps to backend QAListItemDTO
 */
export interface QAListItemDTO {
  id: string;
  question: string;
  answer_preview: string;
  date_from: string;
  date_to: string;
  created_at: string;
  sources_count: number;
}

/**
 * Paginated Q&A list response
 * Maps to backend QAListResponseDTO
 */
export interface QAListResponseDTO {
  items: QAListItemDTO[];
  next_cursor?: string;
  has_more: boolean;
}

/**
 * Error response from API
 */
export interface ErrorResponseDTO {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

/**
 * Form data for Q&A form
 */
export interface QAFormViewModel {
  question: string;
  dateRange: {
    from?: Date;
    to?: Date;
  };
}
