import { createFormHook, createFormHookContexts, type AnyFormApi } from "@tanstack/react-form";

import { FormTextField } from "@/components/form/TextField";
import type { ProblemDetail } from "@/lib/http";

export const { fieldContext, formContext, useFieldContext } = createFormHookContexts();

export const { useAppForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    TextField: FormTextField,
  },
  formComponents: {},
});

export function setProblemErrors(form: AnyFormApi, problem?: ProblemDetail) {
  let fields: Record<string, { message: string }> = {};

  if (problem?.status === 422 && problem.errors) {
    fields = Object.fromEntries(
      Object.entries(problem.errors).map(([field, message]): [string, { message: string }] => [
        field,
        { message: String(message) },
      ]),
    );
  }

  form.setErrorMap({ onSubmit: { fields } });
}
