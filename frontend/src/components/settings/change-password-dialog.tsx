import { useChangePass } from '@/hooks/user/use-change-pass';
import { zodResolver } from '@hookform/resolvers/zod';
import { DialogClose } from '@radix-ui/react-dialog';
import { Check, EyeIcon, EyeOffIcon, Lock, X } from 'lucide-react';
import { useCallback, useState, type ReactNode } from 'react';
import { Controller, useForm, type Control } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import z from 'zod';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import { Field, FieldError, FieldLabel, FieldSet } from '../ui/field';
import { InputGroup, InputGroupAddon, InputGroupInput } from '../ui/input-group';
import { Spinner } from '../ui/spinner';
import { ErrorAlert } from '../view';

const changePasswordSchema = z.object({
  oldPass: z
    .string({
      error: 'SETTINGS.CHANGE_PASS_DIALOG.oldPass_MIN_10',
    })
    .min(10, { error: 'SETTINGS.CHANGE_PASS_DIALOG.oldPass_MIN_10' }),
  newPass: z
    .string({ error: 'SETTINGS.CHANGE_PASS_DIALOG.newPass_MIN_10' })
    .min(10, { error: 'SETTINGS.CHANGE_PASS_DIALOG.newPass_MIN_10' }),
});
type ChangePassFormValues = z.infer<typeof changePasswordSchema>;

export function ChangePasswordDialog({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { control, handleSubmit, getValues } = useForm<ChangePassFormValues>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: {
      oldPass: '',
      newPass: '',
    },
  });

  const { changePass, isPending: loading, error } = useChangePass();
  const [open, setOpen] = useState(false);

  const onChangePass = useCallback(() => {
    return changePass({ data: getValues() }).then(() => {
      setOpen(false);
      toast.success(t('SETTINGS.CHANGE_PASS_DIALOG.CHANGE_PASS_SUCCESS'));
    });
  }, [setOpen, changePass, getValues]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="w-full max-w-4xl">
        <DialogTitle>{t('SETTINGS.CHANGE_PASS_DIALOG.TITLE')}</DialogTitle>
        <DialogDescription>{t('SETTINGS.CHANGE_PASS_DIALOG.DESCRIPTION')}</DialogDescription>
        <ErrorAlert title="ALERT.CHANGE_PASS_ERROR" error={error} />
        <form className="space-y-4">
          <FieldSet>
            <PasswordField control={control} name="oldPass" />
            <PasswordField control={control} name="newPass" />
          </FieldSet>
        </form>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">
              <X />
              {t('ACTION.CANCEL')}
            </Button>
          </DialogClose>

          <Button type="button" onClick={handleSubmit(onChangePass)}>
            {loading ? <Spinner /> : <Check />}
            {t('ACTION.CONFIRM')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function PasswordField({
  control,
  name,
}: {
  control: Control<ChangePassFormValues>;
  name: keyof ChangePassFormValues;
}) {
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const toggleShowPassword = useCallback(() => {
    setShowPassword(!showPassword);
  }, [showPassword, setShowPassword]);

  return (
    <Controller
      name={name}
      control={control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid}>
          <FieldLabel>{t(`SETTINGS.CHANGE_PASS_DIALOG.${name}`)}</FieldLabel>
          <InputGroup>
            <InputGroupInput
              {...field}
              type={showPassword ? 'text' : 'password'}
              autoComplete="off"
              aria-invalid={fieldState.invalid}
            />
            <InputGroupAddon align="inline-start">
              <Lock />
            </InputGroupAddon>
            <InputGroupAddon align="inline-end">
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={toggleShowPassword}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOffIcon /> : <EyeIcon />}
              </Button>
            </InputGroupAddon>
          </InputGroup>
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
