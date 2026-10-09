import { If } from "@samuelsih/reactifx";
import { FormControl, TextField, type TextFieldProps } from "reshaped";

import { useFieldContext } from "@/lib/form";

type Props = Omit<TextFieldProps, "name" | "value" | "defaultValue" | "onChange"> & {
  label: React.ReactNode;
};

export function FormTextField({ label, inputAttributes, ...props }: Props) {
  const field = useFieldContext<string>();
  const isTouched = field.state.meta.isTouched;
  const error = isTouched ? field.state.meta.errors[0]?.message : undefined;

  return (
    <FormControl hasError={Boolean(error)}>
      <FormControl.Label>{label}</FormControl.Label>
      <TextField
        name={field.name}
        value={field.state.value}
        onChange={({ value }) => field.handleChange(value)}
        inputAttributes={{ ...inputAttributes, onBlur: field.handleBlur }}
        {...props}
      />
      <If cond={Boolean(error)}>
        <FormControl.Error>{error}</FormControl.Error>
      </If>
    </FormControl>
  );
}
