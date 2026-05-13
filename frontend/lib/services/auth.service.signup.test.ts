import { vi, beforeEach, afterEach } from "vitest";

// auth.service.signup unit test — RM-M3-01 frontend half (sprint
// claude/work_260513-m). Mocks fetch and asserts the POST /api/v1/auth/signup
// payload + happy path response handling. Full e2e (form submit through to
// HR DB lookup) is carve out — the M3 follow-up sprint's TC-SIGNUP spec.

describe("authService.signup", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    fetchMock = vi.fn();
    globalThis.fetch = fetchMock as unknown as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("posts the payload as JSON to /api/v1/auth/signup and returns parsed body", async () => {
    const { authService } = await import("./auth.service");
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          status: "created",
          data: {
            user_id: "yklee",
            kratos_id: "01H...",
            department: "Engineering",
            message: "Account created successfully. You can now sign in.",
          },
        }),
        { status: 201, headers: { "Content-Type": "application/json" } }
      )
    );

    const result = await authService.signup({
      name: "YK Lee",
      system_id: "yklee",
      employee_id: "1001",
      password: "ChangeMe-12345!",
    });

    expect(result.status).toBe("created");
    expect(result.data.user_id).toBe("yklee");
    expect(result.data.department).toBe("Engineering");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("/api/v1/auth/signup");
    expect(init?.method).toBe("POST");
    expect(init?.headers).toEqual(expect.objectContaining({ "Content-Type": "application/json" }));
    expect(init?.body).toBe(
      JSON.stringify({
        name: "YK Lee",
        system_id: "yklee",
        employee_id: "1001",
        password: "ChangeMe-12345!",
      })
    );
  });

  it("rejects with ApiError on HR DB miss (403)", async () => {
    const { authService } = await import("./auth.service");
    const { ApiError } = await import("./api-client");
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({ status: "forbidden", error: "identity verification failed", code: "hr_lookup_failed" }),
        { status: 403 }
      )
    );

    await expect(
      authService.signup({
        name: "Stranger",
        system_id: "stranger",
        employee_id: "9999",
        password: "Pass-1234567890!",
      })
    ).rejects.toBeInstanceOf(ApiError);
  });
});
