import { If } from "@samuelsih/reactifx";
import { FormControl, NumberField, type NumberFieldProps } from "reshaped";

import { useFieldContext } from "@/lib/form";

type Props = Omit<
  NumberFieldProps,
  "name" | "value" | "defaultValue" | "onChange" | "increaseAriaLabel" | "decreaseAriaLabel"
> & {
  label: string;
};

export function FormNumberField({ label, inputAttributes, ...props }: Props) {
  const field = useFieldContext<number>();
  const isTouched = field.state.meta.isTouched;
  const error = isTouched ? field.state.meta.errors[0]?.message : undefined;

  return (
    <FormControl hasError={Boolean(error)}>
      <FormControl.Label>{label}</FormControl.Label>
      <NumberField
        name={field.name}
        value={field.state.value}
        onChange={({ value }) => field.handleChange(value)}
        increaseAriaLabel={`Increase ${label}`}
        decreaseAriaLabel={`Decrease ${label}`}
        inputAttributes={{ ...inputAttributes, onBlur: field.handleBlur }}
        {...props}
      />
      <If cond={Boolean(error)}>
        <FormControl.Error>{error}</FormControl.Error>
      </If>
    </FormControl>
  );
}
