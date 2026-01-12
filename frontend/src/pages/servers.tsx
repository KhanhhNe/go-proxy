import { Tag } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DataTable, useTable } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { CopyableSpan, CopyTooltip } from "@/components/ui/tooltip";
import {
  cn,
  durationToMs,
  getServerString,
  getTags,
  PROTOCOLS,
  StringableServer,
  useNow,
} from "@/lib/utils";
import {
  findMatchingListener,
  useAppStateStore,
  useManagerStore,
} from "@/state";
import { ManagedProxyServer } from "@bindings/go-proxy";
import {
  DeleteListeners,
  DeleteServers,
  ImportProxyFile,
  ParseProxyLine,
  RecheckServer,
} from "@bindings/go-proxy/myservice";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
  PaginationState,
  RowSelectionState,
} from "@tanstack/react-table";
import saveAs from "file-saver";
import {
  ClipboardIcon,
  DownloadIcon,
  PlusIcon,
  RotateCcw,
  TrashIcon,
  XIcon,
} from "lucide-react";
import { DateTime, Duration } from "luxon";
import { useEffect, useState } from "react";

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
      cell: function CheckCell({ row }) {
        const nowTime = useNow();
        const now = DateTime.fromJSDate(nowTime) as DateTime<true>;
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
              onClick={() => RecheckServer(row.original.Server!.Id)}
            >
              <RotateCcw />
            </Button>

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
  const localIp = useAppStateStore((state) => state.state?.LocalIp);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  });
  const columns = useColumns();
  const deleteServers = useDeleteServers();
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [proxyExportContent, setProxyExportContent] = useState("");
  const previewExportContent = proxyExportContent
    .split("\n")
    .slice(0, 20)
    .join("\n");

  const [importModalOpen, setImportModalOpen] = useState(false);
  const [importContent, setImportContent] = useState("");
  const [separator, setSeparator] = useState(";");
  const [userChangedSeparator, setUserChangedSeparator] = useState(false);
  const [skipCount, setSkipCount] = useState(0);
  const [skipHeader, setSkipHeader] = useState(false);
  const [defaultPort, setDefaultPort] = useState(22);

  const [importPreview, setImportPreview] = useState<StringableServer | null>(
    null,
  );

  const handleImportModalOpenChange = (open: boolean) => {
    setImportModalOpen(open);
    if (!open) {
      setImportContent("");
      setSeparator(";");
      setUserChangedSeparator(false);
      setSkipCount(0);
      setSkipHeader(false);
      setDefaultPort(22);
      setImportPreview(null);
    }
  };

  const detectMostCommonSeparator = (content: string): string => {
    const separators = [";", "|", ":", ","];
    let mostCommon = ";";
    let maxCount = 0;

    for (const sep of separators) {
      const count = content.split(sep).length - 1;
      if (count > maxCount) {
        maxCount = count;
        mostCommon = sep;
      }
    }

    return maxCount > 0 ? mostCommon : ";";
  };

  const handleImportContentChange = (content: string) => {
    setImportContent(content);
    if (!userChangedSeparator && content) {
      setSeparator(detectMostCommonSeparator(content));
    }
  };

  useEffect(() => {
    const lines = importContent.split("\n");
    const line = lines.slice(skipHeader ? 1 : 0)[0];
    if (line) {
      ParseProxyLine(line, separator, skipCount, defaultPort).then(
        setImportPreview,
      );
    } else {
      setTimeout(() => setImportPreview(null));
    }
  }, [defaultPort, importContent, separator, skipCount, skipHeader]);

  const handleImport = () => {
    ImportProxyFile(
      importContent,
      separator,
      skipCount,
      defaultPort,
      skipHeader,
    ).then(() => {
      handleImportModalOpenChange(false);
    });
  };

  const table = useTable({
    data: servers,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onRowSelectionChange: setRowSelection,
    onPaginationChange: setPagination,
    state: {
      rowSelection,
      pagination,
    },
  });

  const selectedServers = Object.keys(rowSelection).map(
    (k) => servers[Number(k)],
  );
  const selectedCount = selectedServers.length;
  const exportingServers = selectedCount > 0 ? selectedServers : servers;

  const showExportModal = (type: "original" | "local") => {
    let res: StringableServer[];

    if (type === "original") {
      res = exportingServers.map((s) => s.Server).filter(Boolean);
    } else {
      res = exportingServers.map((s) => ({
        Host: localIp ?? "localhost",
        Port: s.Server?.Port,
        Protocols: {
          [PROTOCOLS.HTTP]: true,
          [PROTOCOLS.SOCKS5]: true,
        },
        ...s.Server?.Auth,
      }));
    }

    setProxyExportContent(res.map(getServerString).join("\n"));
    setExportModalOpen(true);
  };

  const exportToFile = () => {
    const blob = new Blob([proxyExportContent], {
      type: "text/plain;charset=utf-8",
    });
    saveAs(
      blob,
      `proxies ${exportingServers.length} ${DateTime.now().toLocaleString(DateTime.DATETIME_SHORT)}.txt`,
    );
  };

  const actions = (
    <div className="flex justify-start gap-1">
      <Button onClick={() => handleImportModalOpenChange(true)}>
        <PlusIcon /> Thêm proxy
      </Button>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline">
            <DownloadIcon /> Xuất proxy
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={() => showExportModal("original")}>
            Xuất proxy gốc - toàn bộ
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => showExportModal("original")}
            disabled={selectedCount === 0}
          >
            Xuất proxy gốc - {selectedCount} proxy đã chọn
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => showExportModal("local")}>
            Xuất proxy LAN - toàn bộ
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => showExportModal("local")}
            disabled={selectedCount === 0}
          >
            Xuất proxy LAN - {selectedCount} proxy đã chọn
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Button
        disabled={selectedCount === 0}
        variant="outline"
        onClick={() => deleteServers(selectedServers)}
      >
        <TrashIcon className="text-destructive" /> Xóa {selectedCount} proxy
      </Button>
    </div>
  );

  return (
    <>
      <div>
        <DataTable title="Proxy nguồn" table={table} actions={actions} />
      </div>

      <Dialog
        open={importModalOpen}
        onOpenChange={handleImportModalOpenChange}
        modal
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Nhập proxy</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="import-content" className="text-sm font-medium">
                Nội dung file proxy (
                {importContent.trim().split("\n").filter(Boolean).length -
                  (skipHeader ? 1 : 0)}{" "}
                dòng)
              </Label>
              <Textarea
                id="import-content"
                value={importContent}
                onChange={(e) => handleImportContentChange(e.target.value)}
                placeholder="Dán proxy ở đây"
                rows={8}
              />
            </div>

            <div className="flex gap-4">
              <div className="flex-1">
                <Label htmlFor="separator" className="text-sm font-medium">
                  Phân cách
                </Label>
                <Select
                  value={separator}
                  onValueChange={(v) => {
                    setSeparator(v);
                    setUserChangedSeparator(true);
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>

                  <SelectContent>
                    <SelectItem value=";">;</SelectItem>
                    <SelectItem value="|">|</SelectItem>
                    <SelectItem value=":">:</SelectItem>
                    <SelectItem value=",">,</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex-1">
                <Label htmlFor="skip-count" className="text-sm font-medium">
                  Bỏ qua cột
                </Label>
                <Input
                  id="skip-count"
                  type="number"
                  min={0}
                  value={skipCount}
                  onChange={(e) => setSkipCount(Number(e.target.value))}
                  className="w-full"
                />
              </div>
              <div className="flex-1">
                <Label htmlFor="default-port" className="text-sm font-medium">
                  Port mặc định
                </Label>
                <Input
                  id="default-port"
                  type="number"
                  min={1}
                  value={defaultPort}
                  onChange={(e) => setDefaultPort(Number(e.target.value) || 0)}
                  className="w-full"
                />
              </div>
              <div className="flex flex-col items-start gap-2">
                <Label htmlFor="skip-header" className="text-sm">
                  Bỏ qua dòng đầu
                </Label>
                <Checkbox
                  id="skip-header"
                  checked={skipHeader}
                  onCheckedChange={(v) => setSkipHeader(!!v)}
                />
              </div>
            </div>

            <div className="space-y-3 rounded border bg-secondary/30 p-3">
              <p className="text-sm font-medium">Preview</p>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <Label htmlFor="preview-host" className="text-xs">
                    Host
                  </Label>
                  <Input
                    id="preview-host"
                    type="text"
                    value={importPreview?.Host ?? ""}
                    disabled
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="preview-port" className="text-xs">
                    Port
                  </Label>
                  <Input
                    id="preview-port"
                    type="text"
                    value={importPreview?.Port ?? ""}
                    disabled
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="preview-user" className="text-xs">
                    User
                  </Label>
                  <Input
                    id="preview-user"
                    type="text"
                    value={importPreview?.Auth?.Username ?? ""}
                    disabled
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="preview-pass" className="text-xs">
                    Pass
                  </Label>
                  <Input
                    id="preview-pass"
                    type="text"
                    value={importPreview?.Auth?.Password ?? ""}
                    disabled
                    className="mt-1"
                  />
                </div>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => handleImportModalOpenChange(false)}
            >
              Hủy
            </Button>
            <Button onClick={handleImport}>Nhập</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={exportModalOpen} onOpenChange={setExportModalOpen} modal>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Xuất proxy</DialogTitle>
          </DialogHeader>

          <Textarea
            value={Array(20).fill(previewExportContent).join("\n")}
            rows={15}
            readOnly
            className="overflow-hidden"
            style={{
              maskImage:
                "linear-gradient(to bottom, hsl(var(--background)) 50%, transparent 100%)",
            }}
          />

          <DialogFooter>
            <CopyTooltip
              copyData={[proxyExportContent]}
              onCopy={() => setExportModalOpen(false)}
            >
              <Button>
                <ClipboardIcon /> Sao chép
              </Button>
            </CopyTooltip>
            <Button onClick={exportToFile}>
              <DownloadIcon /> Lưu file
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
