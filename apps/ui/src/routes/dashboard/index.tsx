import { createFileRoute } from "@tanstack/react-router";
import { Text, View } from "reshaped";

export const Route = createFileRoute("/dashboard/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <View padding={6}>
      <Text variant="featured-2" as="h1">
        Dashboard
      </Text>
    </View>
  );
}
