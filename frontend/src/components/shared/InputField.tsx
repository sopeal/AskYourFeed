import { forwardRef } from 'react';
import { Input } from '../ui/input';
import { Label } from '../ui/label';

export interface InputFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  id: string;
}

/**
 * Reusable input field component with label and error display
 * Based on Shadcn/ui Input component
 */
export const InputField = forwardRef<HTMLInputElement, InputFieldProps>(
  ({ label, error, id, className, ...props }, ref) => {
    return (
      <div className="space-y-2">
        <Label htmlFor={id} className="text-sm font-medium">
          {label}
        </Label>
        <Input
          id={id}
          ref={ref}
          className={`${error ? 'border-red-500 focus-visible:ring-red-500' : ''} ${className || ''}`}
          aria-invalid={error ? 'true' : 'false'}
          aria-describedby={error ? `${id}-error` : undefined}
          {...props}
        />
        {error && (
          <p id={`${id}-error`} className="text-sm text-red-500" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  }
);

InputField.displayName = 'InputField';
