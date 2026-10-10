import { If } from "@samuelsih/reactifx";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { Alert, Button, Text, useToggle, View } from "reshaped";
import * as v from "valibot";

import type { LoginRequest } from "@/client/model";
import { RouteLink } from "@/components/RouteLink";
import { setProblemErrors, useAppForm } from "@/lib/form";
import { toProblemDetail, unwrapResponse } from "@/lib/http";
import { authLogin } from "@/client";

const schema = v.object({
  email: v.pipe(
    v.string(),
    v.nonEmpty("Email is required"),
    v.minLength(3, "Email must be at least 3 characters"),
    v.maxLength(320, "Email must be at most 320 characters"),
    v.email("Enter a valid email address"),
  ),
  password: v.pipe(v.string(), v.nonEmpty("Password is required")),
});

export const Route = createFileRoute("/auth/login")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const passwordToggle = useToggle();

  const login = useMutation({
    mutationFn: (value: LoginRequest) => unwrapResponse(authLogin(value)),
    onError: (error) => setProblemErrors(form, toProblemDetail(error)),
    onSuccess: () => navigate({ to: "/dashboard" }),
  });

  const form = useAppForm({
    defaultValues: {
      email: "",
      password: "",
    },
    validators: {
      onChange: schema,
    },
    onSubmit: ({ value, formApi }) => {
      setProblemErrors(formApi);
      login.mutate(value);
    },
  });

  const problem = toProblemDetail(login.error);

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
            Welcome back
          </Text>
          <Text variant="body-2" color="neutral-faded" as="p">
            Log in to your Nibiru account.
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
                label="Password"
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
            <If cond={login.isError && problem.status !== 422}>
              <Alert color="critical" title={problem.title}>
                {problem.detail}
              </Alert>
            </If>
            <Button type="submit" variant="solid" fullWidth loading={login.isPending}>
              Log in
            </Button>
          </View>
        </form>

        <Text variant="body-2" align="center" as="p">
          Don't have an account?{" "}
          <RouteLink to="/auth/register" color="inherit" variant="underline">
            Sign up
          </RouteLink>
        </Text>
      </View>
    </View>
  );
}
