import { useTheme } from '@/hooks/theme-provider';
import { cn } from '@/lib';
import { Check, Copy, X } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import { tomorrow, tomorrowNightBlue } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import { Button } from '../ui/button';

export function ConfigViewer({
  text,
  className,
  onClose,
}: {
  text: string;
  className?: string;
  onClose?: () => void;
}) {
  const { theme } = useTheme();
  const { t } = useTranslation();
  const [style, setStyle] = useState(theme === 'dark' ? tomorrowNightBlue : tomorrow);
  useEffect(() => {
    setStyle(theme === 'dark' ? tomorrowNightBlue : tomorrow);
  }, [theme, setStyle]);

  const [copied, setCopied] = useState(false);

  const copyText = useCallback(() => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 3000);
  }, [setCopied]);

  return (
    <div className={cn('p-4', className)}>
      <div className="rounded-t-lg bg-accent flex items-center">
        {onClose && (
          <Button variant="ghost" size="icon" onClick={onClose} aria-label={t('ACTION.CLOSE')}>
            <X className="size-4" />
          </Button>
        )}
        <span className="text-sm font-medium text-accent-foreground ps-2">config.yaml</span>

        {/* gap between elements */}
        <div className="flex-1"></div>

        <Button
          variant="ghost"
          size={copied ? undefined : 'icon'}
          onClick={copyText}
          aria-label={copied ? t('ALERT.COPIED') : t('ACTION.COPY')}
          className="end-2 top-2 z-10 transform transition-all"
        >
          {copied ? (
            <>
              <Check className="size-4" />
              {t('ALERT.COPIED')}
            </>
          ) : (
            <Copy className="size-4" />
          )}
        </Button>
      </div>

      <SyntaxHighlighter
        language="yaml"
        style={{
          ...style,
          hljs: {
            ...style.hljs,
            background: 'var(--card)',
          },
        }}
        className="rounded-b-lg"
      >
        {text}
      </SyntaxHighlighter>
    </div>
  );
}
