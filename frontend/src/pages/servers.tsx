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
import {
  useAppStateStore,
  useManagerStore,
  useMatchingListener,
} from "@/state";
import { ManagedProxyServer } from "@bindings/go-proxy";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import { Clipboard } from "lucide-react";
import { DateTime, Duration } from "luxon";
import { useMemo, useState } from "react";

export function PageServers() {
  const recheckInterval = useManagerStore(
    (state) => state.manager?.ServerRecheckInterval,
  );
  const servers = useManagerStore((state) =>
    Object.values(state.manager?.Servers || {}).filter(Boolean),
  );
  const localIp = useAppStateStore((s) => s.state?.LocalIp);

  const [rowSelection, setRowSelection] = useState({});

  const columns: ColumnDef<ManagedProxyServer>[] = useMemo(
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
          const now = useNow();
          const lastChecked = DateTime.fromISO(
            row.original.Server?.LastChecked,
          ) as DateTime<true> | null;
          const text =
            lastChecked?.toRelative({
              base: DateTime.fromJSDate(now),
              style: "narrow",
            }) ?? "";

          const recheck = durationToMs(recheckInterval);
          const deadline = recheck
            ? DateTime.now().minus(Duration.fromMillis(recheck))
            : null;

          return (
            <span
              className={cn(
                deadline &&
                  lastChecked &&
                  lastChecked < deadline &&
                  "bg-yellow-400",
              )}
            >
              {text}
            </span>
          );
        },
      },
      {
        id: "listener",
        header: "Port local",
        cell: ({ row }) => {
          const listener = useMatchingListener(row.original.Server?.Id ?? "");

          if (listener) {
            return listener.Listener?.Port;
          }
        },
      },
      {
        id: "actions",
        header: "Hành động",
        cell: ({ row }) => {
          const listener = useMatchingListener(
            row.original.Server?.Id ?? "",
          )?.Listener;

          return (
            <div className="flex gap-1">
              <CopyTooltip copyData={[getServerString(row.original.Server)]}>
                <Button size="icon" variant="outline">
                  <div className="flex flex-col items-center">
                    <span className="text-xs">Proxy</span>
                    <Clipboard />
                  </div>
                </Button>
              </CopyTooltip>

              {listener && (
                <CopyTooltip
                  copyData={[
                    `http://${localIp || "localhost"}:${listener.Port}`,
                  ]}
                >
                  <Button size="icon" variant="outline">
                    <div className="flex flex-col items-center">
                      <span className="text-xs">LAN</span>
                      <Clipboard />
                    </div>
                  </Button>
                </CopyTooltip>
              )}
            </div>
          );
        },
      },
    ],
    [],
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
