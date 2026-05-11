'use client';
import {
  createElement,
  isValidElement,
  memo,
  ReactNode,
  Ref,
  useMemo,
  useRef,
  useState,
} from 'react';
import cx from 'classnames';

import { SwitchAnimation } from '~components/Animations/SwitchAnimation';
import { Skeleton } from '~components/Skeleton/Skeleton';
import { Tooltip } from '~components/Tooltip';
import { Typography } from '~components/Typography';
import { SortIcon } from '~icons/JsxSvg/SortIcon';
import colors from '~styles/colors';

import styles from './styles.module.scss';

export interface Row<TData> {
  id: string;
  index: number;
  original: TData;
}

export interface ColumnDef<TData, TValue = unknown> {
  id?: string;
  header?:
    | ReactNode
    | ((ctx: {
        column: ColumnDef<TData, TValue>;
        table: TableInstance<TData>;
      }) => ReactNode);
  cell?:
    | ReactNode
    | ((
        ctx: CellContext<TData, TValue> & {
          hovered?: boolean;
          additionalData?: any;
        },
      ) => ReactNode);
  accessorKey?: keyof TData | string;
  accessorFn?: (row: TData) => TValue;
}

export interface Cell<TData, TValue = unknown> {
  id: string;
  row: Row<TData>;
  column: ColumnDef<TData, TValue>;
  getValue: () => TValue;
}

export interface TableInstance<TData> {
  rows: Row<TData>[];
  getRowModel: () => { rows: Row<TData>[] };
}

export interface CellContext<TData, TValue = unknown> {
  table: TableInstance<TData>;
  row: Row<TData>;
  column: ColumnDef<TData, TValue>;
  cell: Cell<TData, TValue>;
  getValue: () => TValue;
  renderValue: () => TValue;
}

const flexRender = <TProps extends object>(
  Comp: ReactNode | ((props: TProps) => ReactNode) | undefined,
  props: TProps,
): ReactNode => {
  if (Comp == null) return null;

  if (isValidElement(Comp)) {
    return Comp;
  }

  if (typeof Comp === 'function') {
    return (Comp as (p: TProps) => ReactNode)(props);
  }

  if (typeof Comp === 'object' && (Comp as any).$$typeof) {
    return createElement(Comp as any, props);
  }

  return Comp as ReactNode;
};

export type MarketTableRow =
  | {
      id: string;
      className: string;
    }
  | {
      className: string;
      idArr: string[];
    }[];

const TableRowContent = <T extends object>({
  row,
  table,
  columns,
  markedRow,
  additionalData,
  isLoading,
  haveHovering = true,
}: {
  row: Row<T>;
  table: TableInstance<T>;
  columns: ColumnDef<T>[];
  markedRow?: MarketTableRow;
  additionalData?: any;
  haveHovering?: boolean;
  isLoading?: boolean | string[];
}) => {
  const [hoveredRowId, setHoveredRowId] = useState<string | null>(null);

  const getClassNameRow = (rowId: string) => {
    if (
      typeof markedRow === 'object' &&
      markedRow !== null &&
      !Array.isArray(markedRow)
    ) {
      return rowId === markedRow.id ? `${markedRow.className}` : undefined;
    } else if (Array.isArray(markedRow) && markedRow !== null) {
      return markedRow.reduce((acc, item) => {
        if (item.idArr.includes(rowId)) {
          return `${acc} ${item.className}`;
        }

        return acc;
      }, '');
    } else {
      return undefined;
    }
  };

  return (
    <tr
      className={getClassNameRow(row.id)}
      onMouseEnter={haveHovering ? () => setHoveredRowId(row.id) : undefined}
      onMouseLeave={haveHovering ? () => setHoveredRowId(null) : undefined}
    >
      {columns.map((column, colIndex) => {
        const value =
          typeof column.accessorFn === 'function'
            ? column.accessorFn(row.original)
            : column.accessorKey
              ? (row.original as any)[column.accessorKey]
              : undefined;

        const columnId = column.id ?? String(column.accessorKey ?? colIndex);

        const cell: Cell<T, unknown> = {
          id: `${row.id}_${columnId}`,
          row,
          column,
          getValue: () => value,
        };

        const showSkeleton =
          typeof isLoading === 'boolean'
            ? isLoading
            : !!isLoading?.includes(row.id);

        return (
          <td key={cell.id}>
            <SwitchAnimation switchKey={showSkeleton ? 'skeleton' : 'default'}>
              {showSkeleton ? (
                <Skeleton className={styles.tableCellPlaceholder} />
              ) : column.cell ? (
                flexRender(column.cell, {
                  table,
                  row,
                  column,
                  cell,
                  value,
                  getValue: cell.getValue,
                  renderValue: cell.getValue,
                  hovered: hoveredRowId === row.id,
                  additionalData,
                })
              ) : (
                (value as ReactNode)
              )}
            </SwitchAnimation>
          </td>
        );
      })}
    </tr>
  );
};

const TableRow = memo(TableRowContent) as typeof TableRowContent;

interface ExpandingTableProps<T extends object> {
  data: T[];
  markedRow?: MarketTableRow;
  headers: ColumnDef<T>[];
  className?: string;
  emptyData?: ReactNode;
  sortableCols?: string[];
  isLoading?: boolean | string[];
  onClickSort?: (id?: string) => void;
  sortBy?: string;
  sortType?: 'asc' | 'desc';
  additionalData?: any;
  refBody?: Ref<HTMLTableSectionElement>;
  columnTooltips?: Record<string, ReactNode>;
  haveHovering?: boolean;
}

export const DataTableComponent = <T extends object>({
  data,
  headers: columns,
  className,
  emptyData,
  sortableCols,
  isLoading,
  markedRow,
  onClickSort,
  sortType,
  sortBy,
  additionalData,
  refBody,
  columnTooltips = {},
  haveHovering,
}: ExpandingTableProps<T>) => {
  const tableRef = useRef<HTMLTableElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const rows: Row<T>[] = useMemo(
    () =>
      data.map((item, index) => ({
        id: String(index),
        index,
        original: item,
      })),
    [data],
  );

  const table: TableInstance<T> = useMemo(
    () => ({
      rows,
      getRowModel: () => ({ rows }),
    }),
    [rows],
  );

  const content = useMemo(() => {
    if (rows.length) {
      return (
        <>
          {rows.map((row) => (
            <TableRow
              key={row.id}
              row={row}
              table={table}
              columns={columns}
              isLoading={isLoading}
              additionalData={additionalData}
              markedRow={markedRow}
              haveHovering={haveHovering}
            />
          ))}
        </>
      );
    }

    return (
      <tr>
        <td colSpan={columns.length}>
          {typeof emptyData === 'string' ? (
            <Typography
              customStyles={{
                width: '100%',
                textAlign: 'center',
                display: 'flex',
                justifyContent: 'center',
              }}
              variant="body"
              element="p"
              color={colors['textSecondary']}
              text={`${emptyData}`}
              lineHeight={1.3}
            />
          ) : (
            emptyData
          )}
        </td>
      </tr>
    );
  }, [rows, emptyData, columns, isLoading, additionalData, markedRow, table]);

  return (
    <div
      className={cx(className, styles.baseTable, {
        [styles.emptyTableContainer as string]: !data.length,
      })}
      ref={containerRef}
    >
      <table className={className} ref={tableRef}>
        <thead>
          <tr>
            {columns.map((column, colIndex) => {
              const logicalId =
                column.id ?? String(column.accessorKey ?? colIndex);
              const isSortable = sortableCols?.includes(logicalId);

              return (
                <th key={logicalId} className={styles.headerCell}>
                  <div
                    className={styles.headerCellContent}
                    style={{
                      cursor: isSortable ? 'pointer' : 'default',
                    }}
                    onClick={() => {
                      if (!isSortable) return;
                      onClickSort && onClickSort(logicalId);
                    }}
                  >
                    <Typography>
                      {flexRender(column.header, { column, table })}
                    </Typography>
                    {columnTooltips[logicalId] ? (
                      <div className={styles.tooltipWrapper}>
                        <Tooltip text={String(columnTooltips[logicalId])} />
                      </div>
                    ) : isSortable ? (
                      <div
                        className={styles.ascDescIndicator}
                        data-sorted={
                          sortBy === logicalId ? sortType : undefined
                        }
                      >
                        <SortIcon />
                      </div>
                    ) : null}
                  </div>
                </th>
              );
            })}
          </tr>
        </thead>

        <tbody ref={refBody}>{content}</tbody>
      </table>
    </div>
  );
};

export const DataTable = memo(DataTableComponent) as typeof DataTableComponent;
