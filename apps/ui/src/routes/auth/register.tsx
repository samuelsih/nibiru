import { If } from "@samuelsih/reactifx";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { Alert, Button, Text, useToggle, View } from "reshaped";
import * as v from "valibot";

import type { RegisterRequest } from "@/client/model";
import { authRegister } from "@/client";
import { RouteLink } from "@/components/RouteLink";
import { setProblemErrors, useAppForm } from "@/lib/form";
import { toProblemDetail, unwrapResponse } from "@/lib/http";

const schema = v.object({
  firstName: v.pipe(
    v.string(),
    v.nonEmpty("First name is required"),
    v.minLength(5, "First name must be at least 5 characters"),
    v.maxLength(100, "First name must be at most 100 characters"),
  ),
  lastName: v.pipe(v.string(), v.maxLength(100, "Last name must be at most 100 characters")),
  email: v.pipe(
    v.string(),
    v.nonEmpty("Email is required"),
    v.minLength(3, "Email must be at least 3 characters"),
    v.maxLength(320, "Email must be at most 320 characters"),
    v.email("Enter a valid email address"),
  ),
  password: v.pipe(
    v.string(),
    v.nonEmpty("Password is required"),
    v.minLength(8, "Password must be at least 8 characters"),
    v.maxLength(72, "Password must be at most 72 characters"),
  ),
});

export const Route = createFileRoute("/auth/register")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const passwordToggle = useToggle();

  const register = useMutation({
    mutationFn: (value: RegisterRequest) => unwrapResponse(authRegister(value)),
    onError: (error) => setProblemErrors(form, toProblemDetail(error)),
    onSuccess: () => navigate({ to: "/auth/login" }),
  });

  const form = useAppForm({
    defaultValues: {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
    },
    validators: {
      onChange: schema,
    },
    onSubmit: ({ value, formApi }) => {
      setProblemErrors(formApi);
      register.mutate(value);
    },
  });

  const problem = toProblemDetail(register.error);

  return (
    <View className="layout" direction="column" align="center" justify="center" height="100dvh">
      <View
        direction="column"
        gap={5}
        width="100%"
        maxWidth="380px"
        backgroundColor="elevation-base"
        border
        borderColor="neutral-faded"
        borderRadius="large"
        shadow="raised"
        padding={6}
      >
        <View direction="column" gap={1}>
          <Text variant="featured-3" weight="bold" as="h1">
            Create account
          </Text>
          <Text variant="body-2" color="neutral-faded" as="p">
            Start your Nibiru account.
          </Text>
        </View>

        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            void form.handleSubmit();
          }}
        >
          <form.AppField name="firstName">
            {(field) => <field.TextField label="First Name" placeholder="Jonathan" />}
          </form.AppField>

          <form.AppField name="lastName">
            {(field) => <field.TextField label="Last Name" placeholder="Joestar" />}
          </form.AppField>

          <form.AppField name="email">
            {(field) => (
              <field.TextField
                label="Email"
                placeholder="jonathan@joestar.com"
                inputAttributes={{ type: "email" }}
              />
            )}
          </form.AppField>

          <form.AppField name="password">
            {(field) => (
              <field.TextField
                label="New password"
                placeholder="Password"
                inputAttributes={{
                  type: passwordToggle.active ? "text" : "password",
                }}
                endSlot={
                  <Button
                    variant="ghost"
                    size="small"
                    icon={passwordToggle.active ? EyeOffIcon : EyeIcon}
                    onClick={passwordToggle.toggle}
                  />
                }
              />
            )}
          </form.AppField>

          <View direction="column" gap={3} paddingTop={2}>
            <If cond={register.isError && problem.status !== 422}>
              <Alert color="critical" title={problem.title}>
                {problem.detail}
              </Alert>
            </If>
            <Button type="submit" variant="solid" fullWidth loading={register.isPending}>
              Create account
            </Button>
          </View>
        </form>

        <Text variant="body-2" align="center" as="p">
          Already have an account?{" "}
          <RouteLink to="/auth/login" color="inherit" variant="underline">
            Log in
          </RouteLink>
        </Text>
      </View>
    </View>
  );
}
