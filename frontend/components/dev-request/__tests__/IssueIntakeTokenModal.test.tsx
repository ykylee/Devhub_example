import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { IssueIntakeTokenModal } from "../IssueIntakeTokenModal";
import { devRequestTokenService } from "@/domain/dev-request/service/dev_request_token.service";
import type { IssuedDevRequestIntakeToken } from "@/domain/dev-request/schema/dev_request_token.types";

vi.mock("@/domain/dev-request/service/dev_request_token.service", () => ({
  devRequestTokenService: {
    issue: vi.fn(),
  },
}));

describe("IssueIntakeTokenModal", () => {
  const mockOnClose = vi.fn();
  const mockOnIssued = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("renders form phase initially", () => {
    render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);
    expect(screen.getByRole("heading", { name: /Issue Intake Token/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/Client Label/i)).toBeInTheDocument();
  });

  it("submits form and transitions to reveal phase", async () => {
    const user = userEvent.setup();
    if (!navigator.clipboard) {
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: { writeText: async () => {} },
      });
    }
    const writeTextSpy = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue(undefined);

    const mockToken: IssuedDevRequestIntakeToken = {
      token_id: "tok_123",
      client_label: "test_client",
      source_system: "test_sys",
      allowed_ips: ["1.1.1.1"],
      plain_token: "ptk_abcdef123456",
      created_at: "2026-05-18T12:00:00Z",
      created_by: "system",
      last_used_at: null,
      revoked_at: null,
      expires_at: null,
    };
    vi.mocked(devRequestTokenService.issue).mockResolvedValue(mockToken);
    // PR #637 codex P1 fix: dev 환경 stubEnv → useEffect 가 mount 시 0.0.0.0/0 + ::/0 자동 채움.
    vi.stubEnv("NODE_ENV", "development");
    try {
      render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);
    } finally {
      // stubEnv 는 afterEach 에서 unstubAllEnvs 로 정리 (vitest 자동 처리)
    }

    // useEffect 의 setState 가 hydration 후 적용되므로 waitFor 로 default 검증
    await waitFor(() => {
      const initialIpInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/);
      expect(initialIpInputs[0]).toHaveValue("0.0.0.0/0");
    });

    await user.type(screen.getByLabelText(/Client Label/i), "test_client");
    await user.type(screen.getByLabelText(/Source System/i), "test_sys");
    
    // 2026-06-17 fix: default 가 0.0.0.0/0 + ::/0 pre-filled. user 가 의도적으로
    // 1.1.1.1 만 submit 하려면 default 0.0.0.0/0 을 clear 후 1.1.1.1 입력.
    const ipInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/i);
    await user.clear(ipInputs[0]);
    await user.type(ipInputs[0], "1.1.1.1");

    await user.click(screen.getByRole("button", { name: /Issue Token/i }));

    await waitFor(() => {
      expect(devRequestTokenService.issue).toHaveBeenCalledWith({
        client_label: "test_client",
        source_system: "test_sys",
        // 2026-06-17 fix: default 가 ["0.0.0.0/0", "::/0"] pre-filled. 첫 input 만
        // user 가 clear 후 "1.1.1.1" 입력했으나 두 번째 input (::/0) 는 default 유지.
        allowed_ips: ["1.1.1.1", "::/0"],
      });
    });

    // Expect reveal phase
    expect(await screen.findByText(/Token shown once/i)).toBeInTheDocument();
    
    // Ensure token is hidden by default (dots)
    const codeBlock = screen.getByRole("code");
    expect(codeBlock.textContent).toContain("•".repeat(32));

    // Test Show/Hide toggle
    await user.click(screen.getByRole("button", { name: /Show token/i }));
    expect(codeBlock.textContent).toContain("ptk_abcdef123456");

    await user.click(screen.getByRole("button", { name: /Hide token/i }));
    expect(codeBlock.textContent).toContain("•".repeat(32));

    // Test copy
    await user.click(screen.getByRole("button", { name: /Copy/i }));
    expect(writeTextSpy).toHaveBeenCalledWith("ptk_abcdef123456");

    // Close button in reveal phase
    await user.click(screen.getByRole("button", { name: /저장 완료 — 닫기/i }));
    expect(mockOnClose).toHaveBeenCalled();
  });

  it("prevents ESC key from closing during reveal phase", async () => {
    const user = userEvent.setup();
    const mockToken: IssuedDevRequestIntakeToken = {
      token_id: "tok_123",
      client_label: "test",
      source_system: "sys",
      allowed_ips: ["0.0.0.0/0"],
      plain_token: "ptk_secret",
      created_at: "2026-05-18T12:00:00Z",
      created_by: "system",
      last_used_at: null,
      revoked_at: null,
      expires_at: null,
    };
    vi.mocked(devRequestTokenService.issue).mockResolvedValue(mockToken);
    // PR #637 codex P1 fix: dev 환경 stubEnv → useEffect 가 mount 시 0.0.0.0/0 + ::/0 자동 채움.
    vi.stubEnv("NODE_ENV", "development");
    try {
      render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);
    } finally {
      // stubEnv 는 afterEach 에서 unstubAllEnvs 로 정리 (vitest 자동 처리)
    }

    await waitFor(() => {
      const initialIpInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/);
      expect(initialIpInputs[0]).toHaveValue("0.0.0.0/0");
    });

    // In form phase, ESC should close
    await user.keyboard("{Escape}");
    expect(mockOnClose).toHaveBeenCalledTimes(1);

    mockOnClose.mockClear();

    // Fill and submit to get to reveal phase
    await user.type(screen.getByLabelText(/Client Label/i), "test");
    await user.type(screen.getByLabelText(/Source System/i), "sys");
    // default 가 0.0.0.0/0 + ::/0 (2026-06-17 fix) — IP field 가 pre-filled 이므로
    // 추가 typing 없이 그대로 submit. clear + type 불필요.
    // (placeholder "e.g. 10.0.0.0/24" 와 selector 매치)
    await user.click(screen.getByRole("button", { name: /Issue Token/i }));

    await screen.findByText(/Token shown once/i);

    // In reveal phase, ESC should NOT close
    await user.keyboard("{Escape}");
    expect(mockOnClose).not.toHaveBeenCalled();
  });
  // 2026-06-17 PR #637 codex P1 review 정공법: prod 환경 default = [""] (deny-by-default).
  // dev 환경 (NODE_ENV=development) 에서만 mount 시점에 0.0.0.0/0 + ::/0 자동 채움.
  // vitest 의 NODE_ENV 가 "test" 이므로 본 test 는 (1) prod 정공법 검증 (default = [""])
  // + (2) dev 환경 stub 후 useEffect 적용 검증 (["0.0.0.0/0", "::/0"]).
  it("default allowed_ips = empty array in production env (deny-by-default, codex P1 fix)", () => {
    // vitest 의 NODE_ENV 가 "test" → useEffect 가 분기 적용 안 함 → default [""] 유지.
    render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);
    const ipInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/);
    expect(ipInputs[0]).toHaveValue("");
  });

  it("mount 시 dev 환경 (NODE_ENV=development) 에서 0.0.0.0/0 + ::/0 자동 채움 (2026-06-17 fix)", async () => {
    vi.stubEnv("NODE_ENV", "development");
    try {
      render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);
      // useEffect 의 setState 가 hydration 후 적용되므로 waitFor 로 검증
      await waitFor(() => {
        const ipInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/);
        expect(ipInputs[0]).toHaveValue("0.0.0.0/0");
        expect(ipInputs[1]).toHaveValue("::/0");
      });
    } finally {
      vi.unstubAllEnvs();
    }
  });

});
