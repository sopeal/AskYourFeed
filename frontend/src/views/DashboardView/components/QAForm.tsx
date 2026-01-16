import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Calendar } from 'lucide-react';
import { qaFormSchema, type QAFormData } from '../../../schemas/qa.schema';
import { Label } from '../../../components/ui/label';
import { Textarea } from '../../../components/ui/textarea';
import { SubmitButton } from '../../../components/shared/SubmitButton';

/**
 * Props for QAForm component
 */
interface QAFormProps {
  onSubmit: (question: string, dateFrom?: Date, dateTo?: Date) => void;
  isLoading: boolean;
}

/**
 * QAForm component
 * Form for asking questions with optional date range
 */
export const QAForm = ({ onSubmit, isLoading }: QAFormProps) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<QAFormData>({
    resolver: zodResolver(qaFormSchema),
    defaultValues: {
      question: '',
      dateRange: {
        from: undefined,
        to: undefined,
      },
    },
  });

  /**
   * Handle form submission
   */
  const handleFormSubmit = (data: QAFormData) => {
    onSubmit(data.question, data.dateRange.from, data.dateRange.to);
    // Optionally reset form after submission
    // reset();
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
      {/* Question field */}
      <div className="space-y-2">
        <Label htmlFor="question">
          Pytanie <span className="text-red-500">*</span>
        </Label>
        <Textarea
          id="question"
          placeholder="Np. O czym pisali ludzie, których obserwuję w tym tygodniu?"
          rows={4}
          disabled={isLoading}
          {...register('question')}
          className={errors.question ? 'border-red-500' : ''}
        />
        {errors.question && (
          <p className="text-sm text-red-500">{errors.question.message}</p>
        )}
      </div>

      {/* Date range info - simplified for now */}
      <div className="space-y-2">
        <Label htmlFor="dateRange" className="flex items-center gap-2">
          <Calendar className="w-4 h-4" />
          Zakres dat
        </Label>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Domyślnie: ostatnie 24 godziny
        </p>
        {/* TODO: Implement DateRangePicker component in future iteration */}
        {/* For now, we'll use default date range (last 24 hours) */}
      </div>

      {/* Submit button */}
      <SubmitButton
        isLoading={isLoading}
        loadingText="Generowanie odpowiedzi..."
        className="w-full"
      >
        Zapytaj
      </SubmitButton>

      {/* Character count */}
      <p className="text-xs text-gray-500 dark:text-gray-400 text-right">
        Maksymalnie 2000 znaków
      </p>
    </form>
  );
};
