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
  PROTOCOLS,
  useNow,
} from "@/lib/utils";
import {
  findMatchingListener,
  useAppStateStore,
  useManagerStore,
} from "@/state";
import { ManagedProxyServer } from "@bindings/go-proxy";
import { DeleteListeners, DeleteServers } from "@bindings/go-proxy/myservice";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
  RowSelectionState,
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

const useColumns = () => {
  const { recheckInterval, listeners } = useManagerStore((state) => ({
    recheckInterval: state.manager?.ServerRecheckInterval,
    listeners: state.listeners,
  }));
  const localIp = useAppStateStore((s) => s.state?.LocalIp);
  const deleteServers = useDeleteServers();

  return [
    {
      header: "Host",
      cell: ({ row }) => (
        <CopyableSpan
          text={row.original.Server?.Host}
          contentProps={{ align: "start" }}
        />
      ),
    },
    {
      header: "Port",
      cell: ({ row }) => <CopyableSpan text={row.original.Server?.Port} />,
    },
    {
      header: "User",
      cell: ({ row }) => (
        <CopyableSpan
          text={row.original.Server?.Auth?.Username}
          contentProps={{ align: "start" }}
        />
      ),
    },
    {
      header: "Pass",
      cell: ({ row }) => (
        <CopyableSpan
          text={row.original.Server?.Auth?.Password}
          contentProps={{ align: "start" }}
        />
      ),
    },
    {
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
      header: "Ping",
      cell: ({ row }) => durationToMs(row.original.Server?.Latency),
    },
    {
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
      header: "LAN",
      cell: ({ row }) => {
        const listener = findMatchingListener(
          row.original.Server?.Id ?? "",
          listeners ?? [],
        );

        if (listener) {
          return <CopyableSpan text={listener.Listener?.Port} />;
        }
      },
    },
    {
      header: "Hành động",
      cell: ({ row }) => {
        const listener = findMatchingListener(
          row.original.Server?.Id ?? "",
          listeners ?? [],
        )?.Listener;

        const serverText = getServerString(row.original.Server);
        const lanText = getServerString({
          Host: localIp ?? "localhost",
          Protocols: {
            [PROTOCOLS.HTTP]: true,
            [PROTOCOLS.SOCKS5]: true,
          },
          ...listener,
        });

        return (
          <div className="flex gap-1">
            <CopyTooltip copyData={[serverText]} tooltip={serverText}>
              <Button size="icon" variant="outline">
                <span className="absolute -translate-y-3 text-xs">Gốc</span>
                <ClipboardIcon />
              </Button>
            </CopyTooltip>

            {listener && (
              <CopyTooltip copyData={[lanText]} tooltip={lanText}>
                <Button size="icon" variant="outline">
                  <span className="absolute -translate-y-3 text-xs">LAN</span>
                  <ClipboardIcon />
                </Button>
              </CopyTooltip>
            )}

            <Button
              size="icon"
              variant="outline"
              onClick={() => deleteServers([row.original])}
            >
              <XIcon className="text-destructive" />
            </Button>
          </div>
        );
      },
    },
  ] satisfies ColumnDef<ManagedProxyServer>[];
};

const useDeleteServers = () => {
  const { listeners } = useManagerStore((state) => ({
    listeners: state.listeners,
  }));

  return (servers: ManagedProxyServer[]) => {
    // Find all listeners that are associated with the servers to be deleted
    const listenersToDelete = new Set<number>();
    for (const server of servers) {
      const listener = findMatchingListener(
        server.Server?.Id ?? "",
        listeners ?? [],
      );
      if (listener?.Listener?.Port) {
        listenersToDelete.add(listener.Listener?.Port);
      }
    }

    return Promise.all([
      DeleteServers(servers.map((s) => s.Server?.Id).filter(Boolean)),
      DeleteListeners(Array.from(listenersToDelete).filter(Boolean)),
    ]);
  };
};

export function PageServers() {
  const { servers } = useManagerStore((state) => ({
    servers: state.servers,
  }));
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const columns = useColumns();

  const deleteServers = useDeleteServers();

  const selectedCount = Object.keys(rowSelection).length;

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

  const actions = useMemo(
    () => (
      <div className="flex justify-start gap-1">
        <Button>
          <PlusIcon /> Thêm proxy
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
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

        <Button
          disabled={selectedCount === 0}
          variant="outline"
          onClick={() =>
            deleteServers(
              Object.keys(rowSelection).map((k) => servers[Number(k)]),
            )
          }
        >
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
