export interface ProblemDetail {
  type: string;
  status: number;
  title: string;
  detail: string;
  instance?: string;
  code?: string;
  errors?: Record<string, unknown>;
  [member: string]: unknown;
}

export function toProblemDetail(error: unknown): ProblemDetail {
  if (isProblemDetail(error)) {
    return {
      ...error,
      detail: error.detail ?? "",
    };
  }

  return {
    type: "about:blank",
    status: 500,
    title: "Internal server error, please try again later",
    detail: error instanceof Error ? error.message : "Something went wrong.",
  };
}

type ProblemDetailPayload = Partial<ProblemDetail> & {
  type: string;
  status: number;
  title: string;
};

function isProblemDetail(value: unknown): value is ProblemDetailPayload {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  return (
    "type" in value &&
    typeof value.type === "string" &&
    "status" in value &&
    typeof value.status === "number" &&
    "title" in value &&
    typeof value.title === "string"
  );
}

export async function unwrapResponse<T>(
  response: Promise<{ data: T; status: number }>,
): Promise<T> {
  try {
    const { data, status } = await response;

    if (status >= 400) {
      throw toProblemDetail(data);
    }

    return data;
  } catch (error) {
    throw toProblemDetail(error);
  }
}
