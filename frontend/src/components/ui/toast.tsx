import * as React from "react"
import { X } from "lucide-react"
import { cn } from "@/lib/utils"

export interface ToastProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'destructive' | 'success'
  onClose?: () => void
}

const Toast = React.forwardRef<HTMLDivElement, ToastProps>(
  ({ className, variant = 'default', onClose, children, ...props }, ref) => {
    const variantStyles = {
      default: 'bg-white dark:bg-gray-950 border-gray-200 dark:border-gray-800',
      destructive: 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-900 text-red-900 dark:text-red-100',
      success: 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-900 text-green-900 dark:text-green-100',
    }

    return (
      <div
        ref={ref}
        className={cn(
          "pointer-events-auto flex w-full items-center justify-between gap-4 rounded-lg border p-4 shadow-lg transition-all",
          variantStyles[variant],
          className
        )}
        {...props}
      >
        <div className="flex-1">{children}</div>
        {onClose && (
          <button
            onClick={onClose}
            className="rounded-md p-1 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            aria-label="Zamknij"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    )
  }
)
Toast.displayName = "Toast"

export { Toast }
