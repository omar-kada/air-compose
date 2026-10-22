import { DialogClose } from '@radix-ui/react-dialog';
import { Check, X } from 'lucide-react';
import { useCallback, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import { Spinner } from '../ui/spinner';

export function ConfirmationDialog({
  children,
  onConfirm,
  title,
  description,
}: {
  children?: ReactNode;
  onConfirm: () => Promise<unknown> | undefined;
  title: string;
  description: string;
}) {
  const { t } = useTranslation();

  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleConfirm = useCallback(() => {
    const promise = onConfirm();
    if (promise) {
      setLoading(true);
      promise.then(() => setOpen(false)).finally(() => setLoading(false));
    } else {
      setOpen(false);
    }
  }, [setOpen, onConfirm]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="w-full max-w-4xl">
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>{description}</DialogDescription>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" aria-label={t('ACTION.CANCEL')}>
              <X />
              {t('ACTION.CANCEL')}
            </Button>
          </DialogClose>

          <Button type="button" onClick={handleConfirm} aria-label={t('ACTION.CONFIRM')}>
            {loading ? <Spinner /> : <Check />}
            {t('ACTION.CONFIRM')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
