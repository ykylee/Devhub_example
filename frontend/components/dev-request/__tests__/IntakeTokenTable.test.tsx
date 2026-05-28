import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { IntakeTokenTable } from "@/domain/dev-request/view/IntakeTokenTable";
import type { DevRequestIntakeToken } from "@/lib/services/dev_request_token.types";

describe("IntakeTokenTable", () => {
  const mockTokens: DevRequestIntakeToken[] = [
    {
      token_id: "tok_1",
      client_label: "client-1",
      source_system: "sys-1",
      allowed_ips: ["127.0.0.1/32"],
      created_by: "admin1",
      created_at: "2026-05-15T10:00:00Z",
      revoked_at: null,
      last_used_at: "2026-05-15T10:30:00Z",
      expires_at: null,
    },
    {
      token_id: "tok_2",
      client_label: "client-2",
      source_system: "sys-2",
      allowed_ips: ["0.0.0.0/0"],
      created_by: "admin2",
      created_at: "2026-05-14T10:00:00Z",
      revoked_at: "2026-05-14T11:00:00Z",
      last_used_at: null,
      expires_at: null,
    },
  ];

  it("renders empty state correctly", () => {
    render(<IntakeTokenTable items={[]} onRevoke={vi.fn()} revokingTokenID={null} />);
    expect(screen.getByText("발급된 intake token 이 없습니다")).toBeInTheDocument();
  });

  it("renders active and revoked tokens with correct badges", () => {
    render(<IntakeTokenTable items={mockTokens} onRevoke={vi.fn()} revokingTokenID={null} />);
    
    expect(screen.getByText("client-1")).toBeInTheDocument();
    expect(screen.getByText("client-2")).toBeInTheDocument();

    const activeBadge = screen.getByText("Active");
    expect(activeBadge).toBeInTheDocument();
    expect(activeBadge.className).toContain("text-success");

    const revokedBadge = screen.getByText("Revoked");
    expect(revokedBadge).toBeInTheDocument();
    expect(revokedBadge.className).toContain("text-destructive");
  });

  it("calls onRevoke when revoke button is clicked", async () => {
    const mockOnRevoke = vi.fn();
    const user = userEvent.setup();
    
    render(<IntakeTokenTable items={mockTokens} onRevoke={mockOnRevoke} revokingTokenID={null} />);
    
    // Only one revoke button should be present since tok_2 is already revoked
    const revokeBtn = screen.getByRole("button", { name: /Revoke/i });
    expect(revokeBtn).toBeInTheDocument();

    await user.click(revokeBtn);
    expect(mockOnRevoke).toHaveBeenCalledWith(mockTokens[0]);
  });

  it("disables revoke button when revokingTokenID matches", () => {
    render(<IntakeTokenTable items={mockTokens} onRevoke={vi.fn()} revokingTokenID="tok_1" />);
    
    const revokeBtn = screen.getByRole("button", { name: /Revoking/i });
    expect(revokeBtn).toBeDisabled();
  });
});
