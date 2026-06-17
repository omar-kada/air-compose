import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  getFeaturesQueryOptions,
  getSettingsQueryOptions,
  useFilteredQuery,
  useUpdateSettings,
} from '@/hooks';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCallback, useEffect, type ReactNode } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { ScrollArea } from '../ui/scroll-area';
import { Skeleton } from '../ui/skeleton';
import { ErrorAlert } from '../view';
import { SettingsForm } from './settings-form';
import { formSchema, fromSettings, toSettings, type FormValues } from './settings-form-schema';

export function SettingsSheet({
  children,
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  children?: ReactNode;
}) {
  const { t } = useTranslation();
  const { data: features, error: featuresError } = useFilteredQuery(getFeaturesQueryOptions());
  const disabled = !features?.editSettings;
  const { data: settings, isPending, error } = useFilteredQuery(getSettingsQueryOptions());
  const mergedError = error ?? featuresError;

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    disabled,
  });
  useEffect(() => {
    // init once loaded
    form.reset(fromSettings(settings));
  }, [settings, form]);

  const { updateSettings } = useUpdateSettings();

  const onSubmit = useCallback(() => {
    updateSettings(toSettings(form.getValues()));
    setOpen(false);
  }, [form, setOpen]);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetDescription className="hidden">{t('SETTINGS.DESCRIPTION')}</SheetDescription>
      {children && <SheetTrigger asChild>{children}</SheetTrigger>}
      <SheetContent className="w-full md:w-none flex flex-col h-full">
        <SheetHeader>
          <SheetTitle>{t('SETTINGS.SETTINGS')}</SheetTitle>
          {disabled && (
            <SheetDescription>
              {t('SETTINGS.DISABLED_TITLE')}, {t('SETTINGS.DISABLED_DESCRIPTION')}
            </SheetDescription>
          )}
        </SheetHeader>
        <ScrollArea className="h-1 flex-1 gap-2">
          <div className=" grid auto-rows-min gap-6 px-4 mb-10">
            <ErrorAlert title={t('ALERT.LOAD_SETTINGS_ERROR')} error={mergedError}></ErrorAlert>
            {isPending ? <SettingsSkeleton /> : settings && <SettingsForm form={form} />}
          </div>
        </ScrollArea>
        <SheetFooter>
          <Button type="submit" onClick={form.handleSubmit(onSubmit)}>
            {t('ACTION.SAVE')}
          </Button>
          <SheetClose asChild>
            <Button variant="outline">{t('ACTION.CLOSE')}</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

function SettingsSkeleton() {
  return (
    <div className="flex flex-col w-full">
      <Skeleton className="h-12"></Skeleton>
      <Skeleton className="h-12"></Skeleton>
      <Skeleton className="h-12"></Skeleton>
    </div>
  );
}
