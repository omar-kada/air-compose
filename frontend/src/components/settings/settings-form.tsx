import { useDeleteAccount } from '@/hooks';
import { GitBranch, KeyRound, Timer, Trash, UserIcon, type LucideProps } from 'lucide-react';
import { type ComponentType, type ReactNode } from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Button } from '../ui/button';
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldSeparator,
  FieldSet,
  FieldTitle,
} from '../ui/field';
import { Input } from '../ui/input';
import { ConfirmationDialog, NotificationMultiSelect } from '../view';
import { ChangePasswordDialog } from './change-password-dialog';
import { type FormValues } from './settings-form-schema';

export function SettingsForm({ form }: { form: UseFormReturn<FormValues> }) {
  const { t } = useTranslation();

  const { deleteAccount } = useDeleteAccount();

  return (
    <form>
      <FieldGroup className="mt-2">
        <SettingsSection title={t('SETTINGS.FORM.ACCOUNT')} Icon={UserIcon} className="flex gap-2">
          <ChangePasswordDialog>
            <Button type="button">
              <KeyRound />
              {t('SETTINGS.FORM.CHANGE_PASSWORD')}
            </Button>
          </ChangePasswordDialog>
          <ConfirmationDialog
            title={t('SETTINGS.FORM.DELETE_ACCOUNT_DIALOG.TITLE')}
            description={t('SETTINGS.FORM.DELETE_ACCOUNT_DIALOG.DESCRIPTION')}
            onConfirm={deleteAccount}
          >
            <Button variant="destructive" type="button">
              <Trash />
              {t('SETTINGS.FORM.DELETE_ACCOUNT')}
            </Button>
          </ConfirmationDialog>
        </SettingsSection>
        <SettingsSection title={t('SETTINGS.FORM.GIT')} Icon={GitBranch}>
          <FieldSet>
            <SettingsField form={form} name="repo" />
            <SettingsField form={form} name="branch" />

            <SettingsField form={form} name="username" />
            <SettingsField form={form} name="token" />
          </FieldSet>
        </SettingsSection>

        <SettingsSection title={t('SETTINGS.FORM.AUTO_SYNC')} Icon={Timer}>
          <FieldSet>
            <SettingsField form={form} name="cron" withDescription />
          </FieldSet>
        </SettingsSection>

        <SettingsSection title={t('SETTINGS.FORM.NOTIFICATION')} Icon={Timer}>
          <FieldSet>
            <SettingsField form={form} name="notificationURL" withDescription />
            <NotificationMultiSelect
              name="notificationTypes"
              label="SETTINGS.FORM.notificationTypes"
              form={form}
              withDescription
            ></NotificationMultiSelect>
          </FieldSet>
        </SettingsSection>
      </FieldGroup>
    </form>
  );
}

function SettingsSection({
  title,
  Icon,
  children,
  className,
}: {
  title: string;
  Icon?: ComponentType<LucideProps>;
  children: ReactNode;
  className?: string;
}) {
  return (
    <>
      <FieldSeparator>
        <span className="flex items-center">
          {Icon && <Icon className="size-4 mx-1" />}
          {title}
        </span>
      </FieldSeparator>
      <div className={className}>{children}</div>
    </>
  );
}

function SettingsField({
  form,
  name,
  withDescription = false,
  withPlaceholder = true,
}: {
  form: UseFormReturn<FormValues>;
  name: keyof FormValues;
  withDescription?: boolean;
  withPlaceholder?: boolean;
}) {
  const { t } = useTranslation();
  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid}>
          <FieldTitle>{t(`SETTINGS.FORM.${name}`)}</FieldTitle>
          {withDescription && (
            <FieldDescription>{t(`SETTINGS.FORM.${name}_DESCRIPTION`)}</FieldDescription>
          )}
          <Input
            {...field}
            aria-invalid={fieldState.invalid}
            autoComplete="off"
            placeholder={withPlaceholder ? t(`SETTINGS.FORM.${name}_PLACEHOLDER`) : ''}
          />
          {fieldState.invalid && (
            <FieldError
              errors={[{ ...fieldState.error, message: t(fieldState.error?.message ?? '') }]}
            />
          )}
        </Field>
      )}
    />
  );
}
