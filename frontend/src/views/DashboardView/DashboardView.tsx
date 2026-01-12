import { useState } from 'react';
import { AxiosError } from 'axios';
import { useAskQuestion } from '../../hooks/useAskQuestion';
import { useToast } from '../../hooks/useToast';
import { Toast } from '../../components/ui/toast';
import { QAForm } from './components/QAForm';
import { QAResponse } from './components/QAResponse';
import type { QADetailDTO, CreateQACommand, ErrorResponseDTO } from '../../types/qa.types';

/**
 * DashboardView component
 * Main view for asking questions about the user's feed
 */
export const DashboardView = () => {
  const [lastResponse, setLastResponse] = useState<QADetailDTO | null>(null);
  const { mutate: askQuestion, isPending, data } = useAskQuestion();
  const { toasts, showError, dismissToast } = useToast();

  /**
   * Handle form submission
   * Converts dates to ISO 8601 format and sends to API
   */
  const handleSubmit = (question: string, dateFrom?: Date, dateTo?: Date) => {
    const command: CreateQACommand = {
      question,
      date_from: dateFrom?.toISOString(),
      date_to: dateTo?.toISOString(),
    };

    askQuestion(command, {
      onSuccess: (response) => {
        setLastResponse(response);
      },
      onError: (error: AxiosError<ErrorResponseDTO>) => {
        console.error('Error asking question:', error);
        
        // Show appropriate error message based on status code
        if (error.response?.status === 429) {
          showError('Przekroczono limit zapytań. Spróbuj ponownie później.');
        } else if (error.response?.status && error.response.status >= 500) {
          showError('Wystąpił błąd serwera podczas generowania odpowiedzi.');
        } else if (error.response?.data?.error?.message) {
          showError(error.response.data.error.message);
        } else {
          showError('Wystąpił nieoczekiwany błąd. Spróbuj ponownie.');
        }
      },
    });
  };

  // Use current data if available, otherwise show last successful response
  const displayData = data || lastResponse;

  return (
    <>
      {/* Toast notifications container */}
      <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-md">
        {toasts.map((toast) => (
          <Toast
            key={toast.id}
            variant={toast.variant}
            onClose={() => dismissToast(toast.id)}
          >
            {toast.message}
          </Toast>
        ))}
      </div>

      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid gap-8 lg:grid-cols-3">
          {/* Left column - Form */}
          <div className="lg:col-span-1">
            <div className="bg-white dark:bg-gray-950 rounded-xl border border-gray-200 dark:border-gray-800 p-6 sticky top-8">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                Zadaj pytanie
              </h2>
              <QAForm onSubmit={handleSubmit} isLoading={isPending} />
            </div>
          </div>

          {/* Right column - Response */}
          <div className="lg:col-span-2">
            <QAResponse data={displayData} isLoading={isPending} />
          </div>
        </div>
      </main>
    </div>
    </>
  );
};
