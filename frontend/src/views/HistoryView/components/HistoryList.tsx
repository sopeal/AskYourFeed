import { useState, useEffect, useRef } from 'react';
import { Trash2 } from 'lucide-react';
import { Accordion } from '../../../components/ui/accordion';
import { Button } from '../../../components/ui/button';
import { HistoryItem } from './HistoryItem';
import { DeleteConfirmationDialog } from './DeleteConfirmationDialog';
import type { QAListItemDTO } from '../../../types/qa.types';

interface HistoryListProps {
  items: QAListItemDTO[];
  onDeleteOne: (id: string) => void;
  onDeleteAll: () => void;
  hasNextPage?: boolean;
  isFetchingNextPage?: boolean;
  fetchNextPage?: () => void;
  isDeletingAll?: boolean;
}

/**
 * List of history items with infinite scroll and delete all functionality
 */
export const HistoryList = ({
  items,
  onDeleteOne,
  onDeleteAll,
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
  isDeletingAll = false,
}: HistoryListProps) => {
  const [deleteAllDialogOpen, setDeleteAllDialogOpen] = useState(false);
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null);
  const observerTarget = useRef<HTMLDivElement>(null);

  // Infinite scroll implementation
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage?.();
        }
      },
      { threshold: 0.1 }
    );

    const currentTarget = observerTarget.current;
    if (currentTarget) {
      observer.observe(currentTarget);
    }

    return () => {
      if (currentTarget) {
        observer.unobserve(currentTarget);
      }
    };
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const handleDeleteOne = (id: string) => {
    setDeleteItemId(id);
  };

  const confirmDeleteOne = () => {
    if (deleteItemId) {
      onDeleteOne(deleteItemId);
      setDeleteItemId(null);
    }
  };

  if (items.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">
          Brak historii zapytań. Zadaj pierwsze pytanie!
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Delete All Button */}
      <div className="flex justify-end">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setDeleteAllDialogOpen(true)}
          disabled={isDeletingAll}
          className="text-destructive hover:text-destructive hover:bg-destructive/10"
        >
          <Trash2 className="h-4 w-4 mr-2" />
          {isDeletingAll ? 'Usuwanie...' : 'Usuń wszystko'}
        </Button>
      </div>

      {/* History Items */}
      <Accordion type="single" collapsible className="space-y-2">
        {items.map((item) => (
          <HistoryItem
            key={item.id}
            item={item}
            onDelete={handleDeleteOne}
          />
        ))}
      </Accordion>

      {/* Infinite Scroll Trigger */}
      {hasNextPage && (
        <div ref={observerTarget} className="py-4 text-center">
          {isFetchingNextPage ? (
            <p className="text-sm text-muted-foreground">Ładowanie...</p>
          ) : (
            <p className="text-sm text-muted-foreground">Przewiń, aby załadować więcej</p>
          )}
        </div>
      )}

      {/* Delete All Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteAllDialogOpen}
        onOpenChange={setDeleteAllDialogOpen}
        onConfirm={onDeleteAll}
        title="Usuń całą historię"
        description="Czy na pewno chcesz usunąć całą historię zapytań? Ta operacja jest nieodwracalna."
        isDeleting={isDeletingAll}
      />

      {/* Delete Single Item Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteItemId !== null}
        onOpenChange={(open) => !open && setDeleteItemId(null)}
        onConfirm={confirmDeleteOne}
        title="Usuń wpis"
        description="Czy na pewno chcesz usunąć ten wpis z historii? Ta operacja jest nieodwracalna."
      />
    </div>
  );
};
