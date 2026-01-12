import { AlertCircle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Skeleton } from '../../../components/ui/skeleton';
import { SourceCard } from './SourceCard';
import type { QADetailDTO } from '../../../types/qa.types';

/**
 * Props for QAResponse component
 */
interface QAResponseProps {
  data: QADetailDTO | null;
  isLoading: boolean;
}

/**
 * Parse answer text into bullet points
 * Splits by newlines and filters empty lines
 */
const parseAnswerPoints = (answer: string): string[] => {
  return answer
    .split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0);
};

/**
 * QAResponse component
 * Displays Q&A answer with sources
 * Handles loading, empty, and success states
 */
export const QAResponse = ({ data, isLoading }: QAResponseProps) => {
  // Loading state - show skeleton
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
          </div>
          <div className="space-y-3 mt-6">
            <Skeleton className="h-6 w-24" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
          </div>
        </CardContent>
      </Card>
    );
  }

  // No data state - show empty message
  if (!data) {
    return (
      <Card>
        <CardContent className="py-12">
          <div className="text-center text-gray-500 dark:text-gray-400">
            <p className="text-lg font-medium mb-2">Zadaj pytanie o swój feed</p>
            <p className="text-sm">
              Wpisz pytanie powyżej, aby uzyskać odpowiedź na podstawie postów z X
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Success state - show answer and sources
  const answerPoints = parseAnswerPoints(data.answer);

  return (
    <div className="space-y-6">
      {/* Answer section */}
      <Card>
        <CardHeader>
          <CardTitle>Odpowiedź</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="space-y-2 list-disc list-inside text-gray-700 dark:text-gray-300">
            {answerPoints.map((point, index) => (
              <li key={index} className="leading-relaxed">
                {point}
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      {/* Sources section */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Źródła ({data.sources.length})
        </h3>
        <div className="grid gap-4 md:grid-cols-2">
          {data.sources.map((source) => (
            <SourceCard key={source.x_post_id} source={source} />
          ))}
        </div>
      </div>
    </div>
  );
};
