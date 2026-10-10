import { Outlet, createFileRoute, redirect, useMatchRoute } from "@tanstack/react-router";
import { ContainerIcon, KeyRoundIcon, MenuIcon, XIcon } from "lucide-react";
import { Button, Hidden, MenuItem, Modal, Text, useToggle, View } from "reshaped";

import { authMe } from "@/client";
import { RouteLink } from "@/components/RouteLink";
import { toProblemDetail, unwrapResponse } from "@/lib/http";

export const Route = createFileRoute("/dashboard")({
  beforeLoad: async () => {
    try {
      await unwrapResponse(authMe());
    } catch (error) {
      if (toProblemDetail(error).status !== 401) {
        throw error;
      }

      throw redirect({ to: "/auth/login" });
    }
  },
  component: RouteComponent,
});

function SidebarMenu({ onSelect }: { onSelect?: () => void }) {
  const matchRoute = useMatchRoute();

  return (
    <View className="menu cursor-pointer" direction="column" gap={1}>
      <RouteLink to="/dashboard/sandboxes" variant="plain" color="inherit" onClick={onSelect}>
        <MenuItem
          icon={<ContainerIcon strokeWidth={1.5} />}
          selected={!!matchRoute({ to: "/dashboard/sandboxes" })}
        >
          Sandboxes
        </MenuItem>
      </RouteLink>

      <RouteLink to="/dashboard/api-keys" variant="plain" color="inherit" onClick={onSelect}>
        <MenuItem
          icon={<KeyRoundIcon strokeWidth={1.5} />}
          selected={!!matchRoute({ to: "/dashboard/api-keys" })}
        >
          API Keys
        </MenuItem>
      </RouteLink>
    </View>
  );
}

function RouteComponent() {
  const sidebar = useToggle();

  return (
    <View direction="column" height="100dvh">
      <View
        as="header"
        direction="row"
        align="center"
        justify="space-between"
        height={16}
        paddingInline={4}
        borderBottom
        borderColor="neutral-faded"
        shrink={false}
      >
        <Text variant="featured-5" weight="bold">
          Nibiru
        </Text>

        <Hidden hide={{ l: true }}>
          <Button
            variant="outline"
            icon={MenuIcon}
            attributes={{ "aria-label": "Open navigation" }}
            onClick={sidebar.toggle}
          />
        </Hidden>
      </View>

      <View direction="row" align="stretch" grow overflow="hidden">
        <Hidden hide={{ s: true, m: true, l: false }}>
          <View
            as="aside"
            direction="column"
            width="236px"
            padding={4}
            borderEnd
            borderColor="neutral-faded"
            shrink={false}
          >
            <SidebarMenu />
          </View>
        </Hidden>

        <View grow overflow="auto">
          <Outlet />
        </View>
      </View>

      <Modal
        active={sidebar.active}
        onClose={sidebar.deactivate}
        position="end"
        size="300px"
        ariaLabel="Navigation"
      >
        <View direction="column" gap={2}>
          <View direction="row" justify="end">
            <Button
              variant="ghost"
              size="small"
              icon={XIcon}
              attributes={{ "aria-label": "Close navigation" }}
              onClick={sidebar.deactivate}
            />
          </View>
          <SidebarMenu onSelect={sidebar.deactivate} />
        </View>
      </Modal>
    </View>
  );
}
