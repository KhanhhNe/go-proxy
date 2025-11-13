import { Badge } from "@/components/ui/badge";
import { DataTable, useTable } from "@/components/ui/table";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import { GetManager } from "@wailsjs/go/main/App";
import { main } from "@wailsjs/go/models";
import { useEffect, useMemo, useState } from "react";

const columns: ColumnDef<main.ManagedLocalListener>[] = [
  {
    id: "port",
    header: "Port",
    cell: ({ row }) => <span>{row.original.Listener?.Port}</span>,
  },
  {
    id: "mode",
    header: "Chế độ",
    cell: ({ row }) => {
      const ignoreAll = row.original.Listener?.Filter.IgnoreAll;

      return (
        <Badge variant="outline" className={ignoreAll ? "bg-yellow-400" : ""}>
          {ignoreAll ? "Trực tiếp" : "Proxy"}
        </Badge>
      );
    },
  },
  {
    id: "tags",
    header: "Lọc tags",
    cell: ({ row }) => (
      <div className="flex gap-1">
        {(row.original.Listener?.Filter.Tags ?? []).map((tag) => (
          <Badge>{tag}</Badge>
        ))}
      </div>
    ),
  },
];

export function PageIndex() {
  const [manager, setManager] = useState<main.listenerServerManager | null>(
    null
  );
  const listeners = useMemo(
    () => Object.values(manager?.Listeners || {}),
    [manager]
  );

  const [rowSelection, setRowSelection] = useState({});

  const table = useTable({
    data: listeners,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onRowSelectionChange: setRowSelection,
    state: {
      rowSelection,
    },
  });

  useEffect(() => {
    GetManager().then(setManager);
  }, []);

  return (
    <div>
      <DataTable title="Proxy dùng được" table={table} />
    </div>
  );
}
