import { Tag } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable, useTable } from "@/components/ui/table";
import { CopyableSpan, CopyTooltip } from "@/components/ui/tooltip";
import {
  cn,
  durationToMs,
  getServerString,
  getTags,
  useNow,
} from "@/lib/utils";
import { useManagerStore } from "@/state";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import { main } from "@wailsjs/go/models";
import { Clipboard } from "lucide-react";
import { DateTime } from "luxon";
import { useMemo, useState } from "react";

export function PageServers() {
  const manager = useManagerStore((state) => state.manager);
  const servers = useMemo(
    () => Object.values(manager?.Servers || {}),
    [manager],
  );
  const now = useNow();

  const [rowSelection, setRowSelection] = useState({});

  const columns: ColumnDef<main.ManagedProxyServer>[] = useMemo(
    () => [
      {
        id: "host",
        header: "Host",
        cell: ({ row }) => (
          <CopyableSpan
            text={row.original.Server?.Host}
            contentProps={{ align: "start" }}
          />
        ),
      },
      {
        id: "port",
        header: "Port",
        cell: ({ row }) => <CopyableSpan text={row.original.Server?.Port} />,
      },
      {
        id: "user",
        header: "User",
        cell: ({ row }) => (
          <CopyableSpan
            text={row.original.Server?.Auth?.Username}
            contentProps={{ align: "start" }}
          />
        ),
      },
      {
        id: "password",
        header: "Pass",
        cell: ({ row }) => (
          <CopyableSpan
            text={row.original.Server?.Auth?.Password}
            contentProps={{ align: "start" }}
          />
        ),
      },
      {
        id: "public_ip",
        header: "IP thật",
        cell: ({ row }) => {
          const sameIp =
            row.original.Server?.PublicIp === row.original.Server?.Host;

          return (
            <span className={cn(sameIp && "opacity-25 hover:opacity-100")}>
              <CopyableSpan
                text={row.original.Server?.PublicIp}
                contentProps={{ align: "start" }}
              />
            </span>
          );
        },
      },
      {
        id: "tags",
        header: "Tags",
        cell: ({ row }) => (
          <div className="flex gap-1">
            {getTags(row.original.Tags).map((tag) => (
              <Tag key={tag} text={tag} />
            ))}
          </div>
        ),
      },
      {
        id: "ping",
        header: "Ping",
        cell: ({ row }) => durationToMs(row.original.Server?.Latency),
      },
      {
        id: "lastChecked",
        header: "Check",
        cell: ({ row }) => {
          const since = row.original.Server?.LastChecked;
          return since
            ? DateTime.fromISO(since).toRelative({
                base: DateTime.fromJSDate(now),
                style: "narrow",
              })
            : "";
        },
      },
      {
        id: "actions",
        header: "Hành động",
        cell: ({ row }) => (
          <div className="flex gap-1">
            <CopyTooltip
              copyData={[getServerString(row.original.Server)]}
              triggerProps={{ asChild: true }}
            >
              <Button size="icon" variant="outline">
                <Clipboard />
              </Button>
            </CopyTooltip>
          </div>
        ),
      },
    ],
    [now],
  );

  const table = useTable({
    data: servers,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onRowSelectionChange: setRowSelection,
    state: {
      rowSelection,
    },
  });

  return (
    <div>
      <DataTable title="Proxy nguồn" table={table} />
    </div>
  );
}
