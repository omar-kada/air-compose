// components/grouped-checkboxes.tsx
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib";
import { Check, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "./button";

export type Option = {
  value: string;
  label: string;
};

export type OptionGroup = {
  group: string;
  items: Option[];
};

export type GroupedCheckboxesVariant = "horizontal" | "vertical"
interface GroupedCheckboxesProps {
  variant?: GroupedCheckboxesVariant;
  groups: OptionGroup[];
  value: string[];               // currently selected values
  onChange: (selected: string[]) => void;  // update function
}
interface GroupedCheckboxesProps {
  groups: OptionGroup[];
  value: string[];               // currently selected values (across all groups)
  onChange: (selected: string[]) => void;
}

export function GroupedCheckboxes({ groups, value, onChange, variant = "vertical" }: GroupedCheckboxesProps) {
  value = value ?? []
  const {t} = useTranslation();
  // Helper to toggle a single option
  const handleCheckboxChange = (optionValue: string, checked: boolean) => {
    if (checked) {
      onChange([...value, optionValue]);
    } else {
      onChange(value.filter((v) => v !== optionValue));
    }
  };

  // Select all items in a specific group
  const handleSelectAll = (groupItems: OptionGroup['items']) => {
    const groupValues = groupItems.map(item => item.value);
    // Merge current selection with group values, removing duplicates
    const newSelection = Array.from(new Set([...value, ...groupValues]));
    onChange(newSelection);
  };

  // Clear all items in a specific group
  const handleClearAll = (groupItems: OptionGroup['items']) => {
    const groupValues = groupItems.map(item => item.value);
    // Keep only values that are NOT in this group
    const newSelection = value.filter(v => !groupValues.includes(v));
    onChange(newSelection);
  };

  return (
    <div className="space-y-6">
      {groups.map((group, idx) => {
        const groupValues = group.items.map(item => item.value);
        const allSelected = groupValues.every(v => value.includes(v));
        const noneSelected = groupValues.every(v => !value.includes(v));

        return (
          <div key={idx} className="space-y-2">
            {/* Group header with title and action buttons */}
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium leading-none">{t(group.group)}</h3>
              <div className="space-x-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="xs"
                  onClick={() => handleSelectAll(group.items)}
                  disabled={allSelected}   // disable if already all selected
                >
                    <Check></Check>
                  {t('ACTION.SELECT_ALL')}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="xs"
                  onClick={() => handleClearAll(group.items)}
                  disabled={noneSelected}  // disable if already none selected
                >
                    <X></X>
                  {t('ACTION.CLEAR_ALL')}
                </Button>
              </div>
            </div>

            <Separator className="my-2" />

            {/* Checkbox list for the group */}
            <div className={cn("flex gap-2", variant === "vertical" ? "flex-col": "flex-wrap")}>
              {group.items.map((option) => (
                <div key={option.value} className="flex items-center space-x-2">
                  <Checkbox
                    id={option.value}
                    checked={value.includes(option.value)}
                    onCheckedChange={(checked) =>
                      handleCheckboxChange(option.value, checked === true)
                    }
                  />
                  <Label htmlFor={option.value} className={cn("text-sm font-normal", variant === "vertical" ? "": "text-nowrap")}>
                    {t(option.label)}
                  </Label>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
}