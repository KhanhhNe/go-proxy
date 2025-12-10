import { Tag } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
  ClipboardIcon,
  DownloadIcon,
  PlusIcon,
  TrashIcon,
  XIcon,
} from "lucide-react";
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
  const selectedCount = Object.keys(rowSelection).length;

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
          const now = DateTime.fromJSDate(useNow()) as DateTime<true>;
          const lastChecked = DateTime.fromISO(
            row.original.Server?.LastChecked,
          ) as DateTime<true> | null;

          let text = "";
          if (lastChecked) {
            const start = lastChecked > now ? now : lastChecked;
            text =
              start.toRelative({
                base: now,
                style: "narrow",
              }) ?? "";
          }

          const recheck = durationToMs(recheckInterval);
          const deadline = recheck
            ? now.minus(Duration.fromMillis(recheck))
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
              {text.replaceAll("trước", "")}
            </span>
          );
        },
      },
      {
        id: "listener",
        header: "LAN",
        cell: ({ row }) => {
          const listener = useMatchingListener(row.original.Server?.Id ?? "");

          if (listener) {
            return <CopyableSpan text={listener.Listener?.Port} />;
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
                  <span className="absolute -translate-y-3 text-xs">Gốc</span>
                  <ClipboardIcon />
                </Button>
              </CopyTooltip>

              {listener && (
                <CopyTooltip copyData={[`http://${localIp}:${listener.Port}`]}>
                  <Button size="icon" variant="outline">
                    <span className="absolute -translate-y-3 text-xs">LAN</span>
                    <ClipboardIcon />
                  </Button>
                </CopyTooltip>
              )}

              <Button size="icon" variant="outline">
                <XIcon className="text-destructive" />
              </Button>
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
    getRowId: (row) => row.Server?.Id || "",
    state: {
      rowSelection,
    },
  });

  const actions = useMemo(
    () => (
      <div className="flex justify-start gap-1">
        <Button>
          <PlusIcon /> Thêm proxy
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger>
            <Button variant="outline">
              <DownloadIcon /> Xuất proxy
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem>Xuất proxy gốc - toàn bộ</DropdownMenuItem>
            <DropdownMenuItem disabled={selectedCount === 0}>
              Xuất proxy gốc - {selectedCount} proxy đã chọn
            </DropdownMenuItem>
            <DropdownMenuItem>Xuất proxy LAN - toàn bộ</DropdownMenuItem>
            <DropdownMenuItem disabled={selectedCount === 0}>
              Xuất proxy LAN - {selectedCount} proxy đã chọn
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <Button disabled={selectedCount === 0} variant="outline">
          <TrashIcon className="text-destructive" /> Xóa {selectedCount} proxy
        </Button>
      </div>
    ),
    [selectedCount],
  );

  return (
    <div>
      <DataTable title="Proxy nguồn" table={table} actions={actions} />
    </div>
  );
}
