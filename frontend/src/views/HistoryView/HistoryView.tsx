import { useEffect } from 'react';
import { useHistory } from './hooks/useHistory';
import { HistoryList } from './components/HistoryList';
import { Skeleton } from '../../components/ui/skeleton';
import { useToast } from '../../hooks/useToast';

/**
 * History View - displays paginated list of Q&A history
 * Protected route - requires authentication
 */
export const HistoryView = () => {
  const {
    data,
    isLoading,
    isError,
    error,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    deleteOne,
    deleteAll,
    isDeletingAll,
    deleteError,
  } = useHistory();

  const { toast } = useToast();

  // Show error toast when delete operation fails
  useEffect(() => {
    if (deleteError) {
      toast({
        variant: 'destructive',
        title: 'Błąd',
        description: deleteError.response?.data?.error?.message || 'Nie udało się usunąć. Spróbuj ponownie.',
      });
    }
  }, [deleteError, toast]);

  // Flatten all pages into single array
  const allItems = data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <div className="min-h-screen bg-background">
      <div className="container max-w-4xl mx-auto py-8 px-4">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Historia zapytań</h1>
          <p className="text-muted-foreground">
            Przeglądaj i zarządzaj swoimi poprzednimi pytaniami i odpowiedziami
          </p>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="border rounded-lg p-4">
                <Skeleton className="h-6 w-3/4 mb-2" />
                <Skeleton className="h-4 w-full mb-2" />
                <Skeleton className="h-4 w-2/3" />
              </div>
            ))}
          </div>
        )}

        {/* Error State */}
        {isError && (
          <div className="text-center py-12">
            <p className="text-destructive mb-2">
              Wystąpił błąd podczas ładowania historii
            </p>
            <p className="text-sm text-muted-foreground">
              {error?.response?.data?.error?.message || 'Spróbuj odświeżyć stronę'}
            </p>
          </div>
        )}

        {/* History List */}
        {!isLoading && !isError && (
          <HistoryList
            items={allItems}
            onDeleteOne={deleteOne}
            onDeleteAll={deleteAll}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            fetchNextPage={fetchNextPage}
            isDeletingAll={isDeletingAll}
          />
        )}
      </div>
    </div>
  );
};
