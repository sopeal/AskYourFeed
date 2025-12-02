import { z } from 'zod';

/**
 * Q&A Form Validation Schema
 * Validates user input for asking questions
 */
export const qaFormSchema = z.object({
  question: z
    .string()
    .min(1, 'Pytanie jest wymagane')
    .max(2000, 'Pytanie nie może przekraczać 2000 znaków')
    .trim(),
  dateRange: z.object({
    from: z.date().optional(),
    to: z.date().optional(),
  }).refine(
    (data) => {
      // If both dates are provided, 'from' must be before or equal to 'to'
      if (data.from && data.to) {
        return data.from <= data.to;
      }
      return true;
    },
    {
      message: 'Data początkowa musi być wcześniejsza lub równa dacie końcowej',
      path: ['from'],
    }
  ),
});

/**
 * Type inference from schema
 */
export type QAFormData = z.infer<typeof qaFormSchema>;
