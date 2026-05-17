import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { IssueIntakeTokenModal } from "../IssueIntakeTokenModal";
import { devRequestTokenService } from "@/lib/services/dev_request_token.service";

vi.mock("@/lib/services/dev_request_token.service", () => ({
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

    const mockToken = {
      token_id: "tok_123",
      client_label: "test_client",
      source_system: "test_sys",
      allowed_ips: ["1.1.1.1"],
      plain_token: "ptk_abcdef123456",
    };
    vi.mocked(devRequestTokenService.issue).mockResolvedValue(mockToken as any);

    render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);

    await user.type(screen.getByLabelText(/Client Label/i), "test_client");
    await user.type(screen.getByLabelText(/Source System/i), "test_sys");
    
    // Fill first IP input
    const ipInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/i);
    await user.type(ipInputs[0], "1.1.1.1");

    await user.click(screen.getByRole("button", { name: /Issue Token/i }));

    await waitFor(() => {
      expect(devRequestTokenService.issue).toHaveBeenCalledWith({
        client_label: "test_client",
        source_system: "test_sys",
        allowed_ips: ["1.1.1.1"],
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
    const mockToken = {
      token_id: "tok_123",
      client_label: "test",
      source_system: "sys",
      allowed_ips: ["0.0.0.0/0"],
      plain_token: "ptk_secret",
    };
    vi.mocked(devRequestTokenService.issue).mockResolvedValue(mockToken as any);

    render(<IssueIntakeTokenModal onClose={mockOnClose} onIssued={mockOnIssued} />);

    // In form phase, ESC should close
    await user.keyboard("{Escape}");
    expect(mockOnClose).toHaveBeenCalledTimes(1);

    mockOnClose.mockClear();

    // Fill and submit to get to reveal phase
    await user.type(screen.getByLabelText(/Client Label/i), "test");
    await user.type(screen.getByLabelText(/Source System/i), "sys");
    const ipInputs = screen.getAllByPlaceholderText(/10\.0\.0\.0/i);
    await user.type(ipInputs[0], "0.0.0.0/0");
    await user.click(screen.getByRole("button", { name: /Issue Token/i }));

    await screen.findByText(/Token shown once/i);

    // In reveal phase, ESC should NOT close
    await user.keyboard("{Escape}");
    expect(mockOnClose).not.toHaveBeenCalled();
  });
});
