import { createFormHook, createFormHookContexts } from "@tanstack/react-form";

import { FormTextField } from "@/components/form/TextField";

export const { fieldContext, formContext, useFieldContext } = createFormHookContexts();

export const { useAppForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    TextField: FormTextField,
  },
  formComponents: {},
});
