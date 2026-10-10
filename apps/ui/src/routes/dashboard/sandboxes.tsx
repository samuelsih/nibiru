import { If } from "@samuelsih/reactifx";
import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { PlusIcon, Trash2Icon } from "lucide-react";
import {
  Alert,
  Badge,
  Button,
  Loader,
  Modal,
  Table,
  Text,
  useToggle,
  View,
  type BadgeProps,
} from "reshaped";
import * as v from "valibot";

import { sandboxCreate, sandboxList } from "@/client";
import type { CreateSandboxRequest, SandboxSummaryState } from "@/client/model";
import { setProblemErrors, useAppForm } from "@/lib/form";
import { toProblemDetail, unwrapResponse } from "@/lib/http";

export const Route = createFileRoute("/dashboard/sandboxes")({
  component: RouteComponent,
});

const stateColors = {
  creating: "warning",
  running: "positive",
  stopping: "warning",
  stopped: "neutral",
  starting: "warning",
  deleting: "warning",
  failed: "critical",
} as const satisfies Record<SandboxSummaryState, BadgeProps["color"]>;

const createSandboxSchema = v.object({
  name: v.pipe(v.string(), v.maxLength(255, "Name must be at most 255 characters")),
  cpu: v.pipe(
    v.number(),
    v.integer("CPU must be a whole number"),
    v.minValue(1, "CPU must be at least 1"),
    v.maxValue(32, "CPU must be at most 32"),
  ),
  ram: v.pipe(
    v.number(),
    v.integer("RAM must be a whole number"),
    v.minValue(1, "RAM must be at least 1"),
    v.maxValue(128, "RAM must be at most 128"),
  ),
  disk: v.pipe(
    v.number(),
    v.integer("Disk must be a whole number"),
    v.minValue(1, "Disk must be at least 1"),
    v.maxValue(500, "Disk must be at most 500"),
  ),
});

function CreateSandbox() {
  const queryClient = useQueryClient();
  const modal = useToggle();

  const createSandboxMutation = useMutation({
    mutationFn: (value: CreateSandboxRequest) => unwrapResponse(sandboxCreate(value)),
    onError: (error) => setProblemErrors(form, toProblemDetail(error)),
    onSuccess: () => {
      close();
      void queryClient.invalidateQueries({ queryKey: ["sandboxes"] });
    },
  });
  
  const form = useAppForm({
    defaultValues: { name: "", cpu: 1, ram: 2, disk: 10 },
    validators: {
      onChange: createSandboxSchema,
    },
    onSubmit: ({ value }) => {
      setProblemErrors(form);
      createSandboxMutation.mutate(value);
    },
  });

  const close = () => {
    modal.deactivate();
    createSandboxMutation.reset();
    form.reset();
  };

  const problem = toProblemDetail(createSandboxMutation.error);

  return (
    <>
      <Button icon={PlusIcon} onClick={modal.activate}>
        Create sandbox
      </Button>

      <Modal active={modal.active} onClose={close} padding={6}>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            void form.handleSubmit();
          }}
        >
          <View direction="column" gap={4}>
            <Modal.Title>Create sandbox</Modal.Title>

            <form.AppField name="name">
              {(field) => <field.TextField label="Name" placeholder="worker-1" />}
            </form.AppField>

            <View direction="row" gap={3}>
              <View.Item grow>
                <form.AppField name="cpu">
                  {(field) => <field.NumberField label="CPU (vCPU)" min={1} max={32} />}
                </form.AppField>
              </View.Item>

              <View.Item grow>
                <form.AppField name="ram">
                  {(field) => <field.NumberField label="RAM (GB)" min={1} max={128} />}
                </form.AppField>
              </View.Item>

              <View.Item grow>
                <form.AppField name="disk">
                  {(field) => <field.NumberField label="Disk (GB)" min={1} max={500} />}
                </form.AppField>
              </View.Item>
            </View>

            <If cond={createSandboxMutation.isError && problem.status !== 422}>
              <Alert color="critical" title={problem.title}>
                {problem.detail}
              </Alert>
            </If>

            <View direction="row" justify="end" gap={2}>
              <Button variant="ghost" onClick={close}>
                Cancel
              </Button>
              <Button type="submit" loading={createSandboxMutation.isPending}>
                Create
              </Button>
            </View>
          </View>
        </form>
      </Modal>
    </>
  );
}

function RouteComponent() {
  const sandboxes = useInfiniteQuery({
    queryKey: ["sandboxes"],
    queryFn: ({ pageParam }) =>
      unwrapResponse(sandboxList(pageParam ? { cursor: pageParam } : undefined)),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    retry: false,
  });

  const items = sandboxes.data?.pages.flatMap((page) => page.items) ?? [];
  const problem = toProblemDetail(sandboxes.error);

  return (
    <View padding={6} direction="column" gap={4}>
      <Text variant="featured-4" weight="bold" as="h1">
        Sandboxes
      </Text>

      <View direction="row" justify="start">
        <CreateSandbox />
      </View>

      <If cond={sandboxes.isPending}>
        <View align="center" padding={8}>
          <Loader ariaLabel="Loading sandboxes" />
        </View>
      </If>

      <If cond={sandboxes.isError}>
        <Alert color="critical" title={problem.title}>
          {problem.detail}
        </Alert>
      </If>

      <If cond={sandboxes.isSuccess && items.length === 0}>
        <View align="center" padding={8} border borderColor="neutral-faded" borderRadius="medium">
          <Text color="neutral-faded">No sandboxes yet.</Text>
        </View>
      </If>

      <If cond={items.length > 0}>
        <View direction="column" gap={4}>
          <Table border>
            <Table.Head>
              <Table.Row>
                <Table.Heading>Name</Table.Heading>
                <Table.Heading>State</Table.Heading>
                <Table.Heading align="start">CPU</Table.Heading>
                <Table.Heading align="start">RAM</Table.Heading>
                <Table.Heading align="start">Disk</Table.Heading>
                <Table.Heading align="center">Action</Table.Heading>
              </Table.Row>
            </Table.Head>
            <Table.Body>
              {items.map((sandbox) => (
                <Table.Row key={sandbox.id}>
                  <Table.Cell>
                    <Text weight="medium">{sandbox.name}</Text>
                  </Table.Cell>
                  <Table.Cell>
                    <Badge
                      variant="faded"
                      color={stateColors[sandbox.state]}
                      className="capitalize"
                    >
                      {sandbox.state}
                    </Badge>
                  </Table.Cell>
                  <Table.Cell align="start">{sandbox.cpu} vCPU</Table.Cell>
                  <Table.Cell align="start">{sandbox.ram} GB</Table.Cell>
                  <Table.Cell align="start">{sandbox.disk} GB</Table.Cell>
                  <Table.Cell align="center">
                    <Button
                      variant="ghost"
                      size="small"
                      color="critical"
                      icon={Trash2Icon}
                      attributes={{ "aria-label": `Delete ${sandbox.name}` }}
                    />
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>

          <If cond={sandboxes.hasNextPage}>
            <View align="start">
              <Button
                variant="outline"
                loading={sandboxes.isFetchingNextPage}
                onClick={() => void sandboxes.fetchNextPage()}
              >
                Load more
              </Button>
            </View>
          </If>
        </View>
      </If>
    </View>
  );
}
