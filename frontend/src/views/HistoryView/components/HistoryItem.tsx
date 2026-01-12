import { useState } from 'react';
import { Trash2, ExternalLink } from 'lucide-react';
import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from '../../../components/ui/accordion';
import { Button } from '../../../components/ui/button';
import { Skeleton } from '../../../components/ui/skeleton';
import { useQADetail } from '../hooks/useHistory';
import type { QAListItemDTO } from '../../../types/qa.types';

interface HistoryItemProps {
  item: QAListItemDTO;
  onDelete: (id: string) => void;
}

/**
 * Single history item with accordion for expanding details
 * Lazy loads full answer and sources when expanded
 */
export const HistoryItem = ({ item, onDelete }: HistoryItemProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  
  // Lazy load full details only when expanded
  const { data: detail, isLoading: isLoadingDetail } = useQADetail(item.id, isExpanded);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(item.id);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('pl-PL', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  return (
    <AccordionItem value={item.id} className="border rounded-lg px-4 mb-2">
      <div className="flex items-start justify-between gap-4">
        <AccordionTrigger
          className="flex-1 text-left hover:no-underline py-4"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          <div className="flex-1 pr-4">
            <h3 className="font-semibold text-lg mb-1">{item.question}</h3>
            <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
              {item.answer_preview}
            </p>
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <span>{formatDate(item.created_at)}</span>
              <span>•</span>
              <span>{item.sources_count} {item.sources_count === 1 ? 'źródło' : 'źródeł'}</span>
            </div>
          </div>
        </AccordionTrigger>
        
        <Button
          variant="ghost"
          size="icon"
          onClick={handleDelete}
          className="mt-4 text-destructive hover:text-destructive hover:bg-destructive/10"
          title="Usuń wpis"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <AccordionContent className="pb-4">
        {isLoadingDetail ? (
          <div className="space-y-3">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
          </div>
        ) : detail ? (
          <div className="space-y-4">
            {/* Full Answer */}
            <div>
              <h4 className="font-semibold mb-2">Odpowiedź:</h4>
              <p className="text-sm whitespace-pre-wrap">{detail.answer}</p>
            </div>

            {/* Date Range */}
            <div className="text-xs text-muted-foreground">
              <span>Zakres dat: </span>
              <span>{formatDate(detail.date_from)} - {formatDate(detail.date_to)}</span>
            </div>

            {/* Sources */}
            {detail.sources && detail.sources.length > 0 && (
              <div>
                <h4 className="font-semibold mb-3">Źródła ({detail.sources.length}):</h4>
                <div className="space-y-3">
                  {detail.sources.map((source) => (
                    <div
                      key={source.x_post_id}
                      className="border rounded-lg p-3 bg-muted/30"
                    >
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div className="flex-1">
                          <div className="font-medium text-sm">
                            {source.author_display_name}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            @{source.author_handle}
                          </div>
                        </div>
                        <a
                          href={source.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:text-primary/80"
                          title="Otwórz post"
                        >
                          <ExternalLink className="h-4 w-4" />
                        </a>
                      </div>
                      <p className="text-sm whitespace-pre-wrap">
                        {source.text || source.text_preview}
                      </p>
                      <div className="text-xs text-muted-foreground mt-2">
                        {formatDate(source.published_at)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            Nie udało się załadować szczegółów.
          </p>
        )}
      </AccordionContent>
    </AccordionItem>
  );
};
