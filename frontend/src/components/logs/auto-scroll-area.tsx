import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib';
import { ChevronDown } from 'lucide-react';
import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';

interface AutoScrollAreaProps {
  children?: ReactNode;
  className?: string; // for the outer container
  watch?: unknown;
}

export function AutoScrollArea({ children, className = '', watch }: AutoScrollAreaProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const isBottomVisibleRef = useRef(true);
  const [showResetButton, setShowResetButton] = useState(false);
  const userHasScrolledRef = useRef(false);

  const setUserHasScrolled = useCallback(
    (value: boolean) => {
      userHasScrolledRef.current = value;
      setShowResetButton(value);
    },
    [userHasScrolledRef, setShowResetButton],
  );

  // Helper to scroll smoothly to bottom
  const scrollToBottom = useCallback(
    (behavior: ScrollBehavior = 'instant') => {
      setUserHasScrolled(false);
      bottomRef.current?.scrollIntoView({ behavior });
    },
    [setUserHasScrolled, userHasScrolledRef],
  );

  const smoothScrollToBottom = useCallback(() => scrollToBottom('smooth'), []);

  // 1. Listen for explicit user scroll inputs (wheel, touch, key navigation)
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const onUserScrollIntent = () => {
      if (!userHasScrolledRef.current && !isBottomVisibleRef.current) {
        setUserHasScrolled(true);
      } else if (userHasScrolledRef.current && isBottomVisibleRef.current) {
        setUserHasScrolled(false);
      }
    };

    const onKeyDown = (e: KeyboardEvent) => {
      const scrollKeys = ['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End', ' '];
      if (scrollKeys.includes(e.key)) {
        onUserScrollIntent();
      }
    };

    container.addEventListener('wheel', onUserScrollIntent, { passive: true });
    container.addEventListener('touchmove', onUserScrollIntent, { passive: true });
    container.addEventListener('keydown', onKeyDown);

    scrollToBottom();

    return () => {
      container.removeEventListener('wheel', onUserScrollIntent);
      container.removeEventListener('touchmove', onUserScrollIntent);
      container.removeEventListener('keydown', onKeyDown);
    };
  }, []);

  useEffect(() => {
    if (!userHasScrolledRef.current) {
      scrollToBottom();
    }
  }, [watch, userHasScrolledRef]);

  // Observe the bottom anchor with IntersectionObserver
  useEffect(() => {
    const bottomElement = bottomRef.current;
    if (!bottomElement) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        const isIntersecting = entry.isIntersecting;
        isBottomVisibleRef.current = isIntersecting;
      },
      { root: null, threshold: 0.2 },
    );

    observer.observe(bottomElement);
    return () => observer.disconnect();
  }, []);

  return (
    <div className={cn('relative flex flex-col', className)}>
      {/* ScrollArea now uses h-full – it fills the parent */}
      <ScrollArea className="w-full h-full" ref={containerRef}>
        <div className="p-4">
          {children}
          <div ref={bottomRef} className="h-1" />
        </div>
      </ScrollArea>

      {/* Floating button – fixed to viewport */}
      {showResetButton && (
        <Button
          aria-label="Scroll to bottom"
          onClick={smoothScrollToBottom}
          className="fixed bottom-16 right-6 rounded-full p-3 shadow-lg z-50"
          size="icon"
        >
          <ChevronDown className="h-5 w-5" />
        </Button>
      )}
    </div>
  );
}
