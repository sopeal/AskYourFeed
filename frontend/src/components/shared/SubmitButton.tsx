import { Button } from '../ui/button';
import { Loader2 } from 'lucide-react';

export interface SubmitButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading?: boolean;
  loadingText?: string;
  children: React.ReactNode;
}

/**
 * Reusable submit button component with loading state
 * Based on Shadcn/ui Button component
 */
export const SubmitButton = ({
  isLoading = false,
  loadingText = 'Ładowanie...',
  children,
  disabled,
  className,
  ...props
}: SubmitButtonProps) => {
  return (
    <Button
      type="submit"
      disabled={isLoading || disabled}
      className={className}
      {...props}
    >
      {isLoading ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {loadingText}
        </>
      ) : (
        children
      )}
    </Button>
  );
};
