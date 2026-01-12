import { ExternalLink, User } from 'lucide-react';
import type { QASourceDTO } from '../../../types/qa.types';

/**
 * Props for SourceCard component
 */
interface SourceCardProps {
  source: QASourceDTO;
}

/**
 * Format date to Polish locale
 */
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('pl-PL', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
};

/**
 * SourceCard component
 * Displays a single source (X post) with author, content, and link
 */
export const SourceCard = ({ source }: SourceCardProps) => {
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      {/* Author info */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
            <User className="w-4 h-4 text-gray-600 dark:text-gray-400" />
          </div>
          <div>
            <p className="font-semibold text-sm text-gray-900 dark:text-gray-100">
              {source.author_display_name}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400">
              @{source.author_handle}
            </p>
          </div>
        </div>
        
        {/* Link to original post */}
        <a
          href={source.url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
          aria-label="Otwórz oryginalny post"
        >
          <ExternalLink className="w-4 h-4" />
        </a>
      </div>

      {/* Post content */}
      <p className="text-sm text-gray-700 dark:text-gray-300 mb-3 whitespace-pre-wrap">
        {source.text_preview || source.text}
      </p>

      {/* Published date */}
      <p className="text-xs text-gray-500 dark:text-gray-400">
        {formatDate(source.published_at)}
      </p>
    </div>
  );
};
