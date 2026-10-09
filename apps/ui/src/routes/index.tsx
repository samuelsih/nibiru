import { createFileRoute } from "@tanstack/react-router";
import { Button, Text, View } from "reshaped";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
  return (
    <View
      className="layout"
      direction="column"
      align="center"
      justify="center"
      gap={4}
      textAlign="center"
      height="100dvh"
      overflow="hidden"
    >
      <View position="absolute" insetTop={3} insetEnd={0}>
        <Button variant="solid" size="small" href="/dashboard">
          Dashboard
        </Button>
      </View>

      <Text variant="headline-1" weight="bold" as="h1">
        Nibiru
      </Text>

      <Text variant="featured-3" as="p">
        <View as="span" backgroundColor="black" paddingInline={2}>
          Open source
        </View>{" "}
        sandbox platform
      </Text>

      <Text variant="body-1" as="p" className="max-w-xl" wrap="balance">
        Spin up isolated sandboxes for your agents and code, with persistent files and a simple API,
        all running on infrastructure you own.
      </Text>
    </View>
  );
}
