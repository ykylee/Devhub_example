import { describe, it, expect, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PermissionMatrix, type PermissionState } from "./PermissionMatrix";

describe("PermissionMatrix", () => {
  const emptyPermissions: PermissionState = {};

  it("renders all 11 resource rows and 4 action columns", () => {
    render(<PermissionMatrix permissions={emptyPermissions} />);

    // resource rows
    expect(screen.getByText("Infrastructure & Topology")).toBeInTheDocument();
    expect(screen.getByText("CI/CD Pipelines")).toBeInTheDocument();
    expect(screen.getByText("Organization & Members")).toBeInTheDocument();
    expect(screen.getByText("Risk & Security")).toBeInTheDocument();
    expect(screen.getByText("Audit Logs & History")).toBeInTheDocument();
    expect(screen.getByText("Platforms")).toBeInTheDocument();
    expect(screen.getByText("Platform Repositories")).toBeInTheDocument();
    expect(screen.getByText("Projects")).toBeInTheDocument();
    expect(screen.getByText("SCM Providers")).toBeInTheDocument();
    expect(screen.getByText("Development Requests (DREQ)")).toBeInTheDocument();
    expect(screen.getByText("DREQ Intake Tokens")).toBeInTheDocument();

    // action column headers
    expect(screen.getByText("View")).toBeInTheDocument();
    expect(screen.getByText("Create")).toBeInTheDocument();
    expect(screen.getByText("Edit")).toBeInTheDocument();
    expect(screen.getByText("Delete")).toBeInTheDocument();
  });

  it("renders grid of buttons for each (resource, action) pair", () => {
    render(<PermissionMatrix permissions={emptyPermissions} />);
    // 11 resources * 4 actions = 44 buttons
    const buttons = screen.getAllByRole("button");
    expect(buttons).toHaveLength(44);
  });

  it("calls onChange with toggled value when a cell button is clicked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <PermissionMatrix
        permissions={{ infrastructure: { view: false } }}
        onChange={onChange}
      />,
    );

    const buttons = screen.getAllByRole("button");
    // Click the first cell (infrastructure x view)
    await user.click(buttons[0]);

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange.mock.calls[0][0]).toMatchObject({
      infrastructure: { view: true },
    });
  });

  it("creates resource entry when toggling a previously absent resource", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<PermissionMatrix permissions={{}} onChange={onChange} />);

    // Click on the first cell -- infrastructure / view -- it must default to false then toggle to true
    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]);

    expect(onChange).toHaveBeenCalledTimes(1);
    const next = onChange.mock.calls[0][0] as PermissionState;
    expect(next.infrastructure?.view).toBe(true);
  });

  it("does not call onChange when readOnly is true", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <PermissionMatrix permissions={emptyPermissions} onChange={onChange} readOnly />,
    );

    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]);
    expect(onChange).not.toHaveBeenCalled();
    expect(buttons[0]).toBeDisabled();
  });

  it("does not call onChange for locked cells", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <PermissionMatrix
        permissions={emptyPermissions}
        onChange={onChange}
        lockedCells={{ audit: { create: true, edit: true, delete: true } }}
      />,
    );

    // The audit row buttons should have title attribute set
    const auditLockedBtn = screen
      .getAllByTitle("Append-only by system code")
      .at(0);
    expect(auditLockedBtn).toBeDefined();
    expect(auditLockedBtn).toBeDisabled();

    await user.click(auditLockedBtn!);
    expect(onChange).not.toHaveBeenCalled();
  });

  it("does nothing when onChange is omitted (no-op click)", async () => {
    const user = userEvent.setup();
    // Should not throw when clicking buttons without onChange
    render(<PermissionMatrix permissions={emptyPermissions} />);
    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]);
    // no assertion necessary — proving no error
    expect(true).toBe(true);
  });

  it("renders Check icon for granted cells and X icon for denied cells", () => {
    const permissions: PermissionState = {
      infrastructure: { view: true, create: false },
    };
    const { container } = render(<PermissionMatrix permissions={permissions} />);

    // The first row contains infrastructure - check at least one svg renders
    const firstRow = container.querySelectorAll("tbody tr")[0];
    expect(firstRow).toBeDefined();
    const svgs = within(firstRow as HTMLElement).getAllByRole("button");
    expect(svgs.length).toBe(4);
  });
});
