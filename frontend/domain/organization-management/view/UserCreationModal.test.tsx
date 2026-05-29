import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

vi.mock("framer-motion", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };
  const motion = new Proxy(
    {},
    {
      get: (_target, tag) =>
        ({ children, ...props }: AnyProps) =>
          React.createElement(tag as string, props, children),
    },
  );
  return {
    motion,
    AnimatePresence: ({ children }: AnyProps) =>
      React.createElement(React.Fragment, null, children),
  };
});

const lookupHR = vi.fn();
const createUser = vi.fn();
vi.mock("@/domain/organization-management/service/identity.service", () => ({
  identityService: {
    lookupHR: (...args: unknown[]) => lookupHR(...args),
    createUser: (...args: unknown[]) => createUser(...args),
  },
}));

import { UserCreationModal } from "./UserCreationModal";
import type { OrgMember } from "@/domain/organization-management/service/identity.service";
import type { Role } from "@/domain/rbac-permissions/schema/rbac.types";

const roles: Role[] = [
  { id: 1, name: "Developer", description: "" },
  { id: 2, name: "Manager", description: "" },
  { id: 3, name: "System Admin", description: "" },
];

const sampleMember: OrgMember = {
  id: "yklee",
  name: "YK Lee",
  email: "yklee@example.com",
  primary_dept_id: "d1",
  current_dept_id: "d1",
  is_seconded: false,
  role: "Developer",
  status: "active",
  appointments: [],
  joined_at: "",
};

beforeEach(() => {
  lookupHR.mockReset();
  createUser.mockReset();
});

describe("UserCreationModal", () => {
  it("renders Human / System toggle buttons", () => {
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    expect(screen.getByText("Human")).toBeInTheDocument();
    expect(screen.getByText("System/AI")).toBeInTheDocument();
  });

  it("switches to system type when System/AI button is clicked", () => {
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    fireEvent.click(screen.getByText("System/AI"));
    // System account changes Create button label
    expect(screen.getByRole("button", { name: /Create System Account/i })).toBeInTheDocument();
  });

  it("succeeds HR lookup and populates email", async () => {
    lookupHR.mockResolvedValueOnce({ email: "lookup@example.com" });
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const userIdInput = screen.getByPlaceholderText(/yklee/);
    fireEvent.change(userIdInput, { target: { value: "yklee" } });
    fireEvent.click(screen.getByTitle("Fetch from HR DB"));
    await waitFor(() => {
      const emailInput = screen.getByPlaceholderText(
        /gardener@devhub\.internal/,
      ) as HTMLInputElement;
      expect(emailInput.value).toBe("lookup@example.com");
    });
  });

  it("shows error when HR lookup fails", async () => {
    lookupHR.mockRejectedValueOnce(new Error("not found"));
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    const userIdInput = screen.getByPlaceholderText(/yklee/);
    fireEvent.change(userIdInput, { target: { value: "yklee" } });
    fireEvent.click(screen.getByTitle("Fetch from HR DB"));
    await waitFor(() => {
      expect(screen.getByText(/Not found in HR database/)).toBeInTheDocument();
    });
  });

  it("does nothing when lookup is clicked with empty user_id", () => {
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    // Button is disabled when user_id empty
    const btn = screen.getByTitle("Fetch from HR DB");
    expect(btn).toBeDisabled();
  });

  it("submits createUser and calls onCreated + onClose on success", async () => {
    createUser.mockResolvedValueOnce(sampleMember);
    const onCreated = vi.fn();
    const onClose = vi.fn();
    const { container } = render(
      <UserCreationModal roles={roles} onClose={onClose} onCreated={onCreated} />,
    );

    fireEvent.change(screen.getByPlaceholderText(/yklee/), {
      target: { value: "yklee" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. YK Lee"), {
      target: { value: "YK Lee" },
    });
    fireEvent.change(screen.getByPlaceholderText(/gardener@devhub\.internal/), {
      target: { value: "yk@example.com" },
    });

    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createUser).toHaveBeenCalledWith({
        user_id: "yklee",
        email: "yk@example.com",
        display_name: "YK Lee",
        role: "Developer",
        status: "active",
        type: "human",
      });
    });
    expect(onCreated).toHaveBeenCalledWith(sampleMember);
    expect(onClose).toHaveBeenCalled();
  });

  it("shows error when createUser fails", async () => {
    createUser.mockRejectedValueOnce(new Error("conflict"));
    const { container } = render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );

    fireEvent.change(screen.getByPlaceholderText(/yklee/), {
      target: { value: "yk" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. YK Lee"), {
      target: { value: "YK" },
    });
    fireEvent.change(screen.getByPlaceholderText(/gardener@devhub\.internal/), {
      target: { value: "yk@example.com" },
    });
    const form = container.querySelector("form");
    if (!form) throw new Error("form not found");
    fireEvent.submit(form);

    await waitFor(() => {
      expect(screen.getByText("conflict")).toBeInTheDocument();
    });
  });

  it("renders mock HR personnel buttons", () => {
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    expect(screen.getByText("akim")).toBeInTheDocument();
    expect(screen.getByText("sjones")).toBeInTheDocument();
  });

  it("populates user_id when mock HR personnel button is clicked", () => {
    render(
      <UserCreationModal roles={roles} onClose={vi.fn()} onCreated={vi.fn()} />,
    );
    fireEvent.click(screen.getByText("akim"));
    const userIdInput = screen.getByPlaceholderText(/yklee/) as HTMLInputElement;
    expect(userIdInput.value).toBe("akim");
  });

  it("calls onClose when Escape is pressed", () => {
    const onClose = vi.fn();
    render(
      <UserCreationModal roles={roles} onClose={onClose} onCreated={vi.fn()} />,
    );
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when Cancel button is clicked", () => {
    const onClose = vi.fn();
    render(
      <UserCreationModal roles={roles} onClose={onClose} onCreated={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onClose).toHaveBeenCalled();
  });
});
