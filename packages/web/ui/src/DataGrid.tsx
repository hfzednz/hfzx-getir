import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

/** Row shape for dense admin tables — plain objects (index signature optional). */
export type DataGridRow = object;

export interface DataGridColumnDef<T extends DataGridRow = DataGridRow> {
  id: string;
  header: ReactNode;
  accessorKey?: keyof T & string;
  cell?: (ctx: { row: T; value: unknown; rowIndex: number }) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string | number;
  className?: string;
}

export interface DataGridProps<T extends DataGridRow = DataGridRow>
  extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
  columns: DataGridColumnDef<T>[];
  data: T[];
  getRowId?: (row: T, index: number) => string;
  emptyMessage?: ReactNode;
  stickyHeader?: boolean;
  onRowClick?: (row: T, index: number) => void;
}

function cellValue<T extends DataGridRow>(
  row: T,
  col: DataGridColumnDef<T>,
): unknown {
  if (col.accessorKey) {
    return (row as Record<string, unknown>)[col.accessorKey];
  }
  return undefined;
}

export function DataGrid<T extends DataGridRow = DataGridRow>({
  columns,
  data,
  getRowId,
  emptyMessage = "No results",
  stickyHeader = true,
  onRowClick,
  className,
  ...props
}: DataGridProps<T>) {
  return (
    <div className={cn("nx-data-grid", className)} {...props}>
      <table className="nx-data-grid__table">
        <thead>
          <tr className="nx-data-grid__thead-row">
            {columns.map((col) => (
              <th
                key={col.id}
                scope="col"
                className={cn(
                  "nx-data-grid__th",
                  stickyHeader && "nx-data-grid__th--sticky",
                  col.align === "center" && "nx-data-grid__th--center",
                  col.align === "right" && "nx-data-grid__th--right",
                  col.className,
                )}
                style={col.width != null ? { width: col.width } : undefined}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="nx-data-grid__empty">
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row, rowIndex) => {
              const id = getRowId?.(row, rowIndex) ?? String(rowIndex);
              return (
                <tr
                  key={id}
                  className={cn(
                    "nx-data-grid__row",
                    onRowClick && "nx-data-grid__row--clickable",
                  )}
                  onClick={onRowClick ? () => onRowClick(row, rowIndex) : undefined}
                >
                  {columns.map((col) => {
                    const value = cellValue(row, col);
                    const content = col.cell
                      ? col.cell({ row, value, rowIndex })
                      : (value as ReactNode);
                    return (
                      <td
                        key={col.id}
                        className={cn(
                          "nx-data-grid__td",
                          col.align === "center" && "nx-data-grid__td--center",
                          col.align === "right" && "nx-data-grid__td--right",
                          col.className,
                        )}
                      >
                        {content ?? "—"}
                      </td>
                    );
                  })}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
