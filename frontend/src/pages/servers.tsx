import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable, useTable } from "@/components/ui/table";
import {
  CopyableSpan,
  CopyTooltip,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import countries from "@/lib/countries.json";
import { cn, getServerString, getTags } from "@/lib/utils";
import { useManagerStore } from "@/state";
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import { main } from "@wailsjs/go/models";
import { Clipboard } from "lucide-react";
import { useMemo, useState } from "react";

const columns: ColumnDef<main.ManagedProxyServer>[] = [
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
        {getTags(row.original.Tags).map((tag) => {
          const country = countries.find((c) => c.code === tag.toUpperCase());

          return (
            <Tooltip key={tag}>
              <TooltipTrigger>
                <Badge key={tag}>
                  {country && (
                    <div className={`flag:${tag} mr-1 rounded`}></div>
                  )}
                  {tag}
                </Badge>
              </TooltipTrigger>
              {country && (
                <TooltipContent>
                  <p>{country.vietnamese_name}</p>
                </TooltipContent>
              )}
            </Tooltip>
          );
        })}
      </div>
    ),
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
];

export function PageServers() {
  const manager = useManagerStore((state) => state.manager);
  const servers = useMemo(
    () => Object.values(manager?.Servers || {}),
    [manager],
  );

  const [rowSelection, setRowSelection] = useState({});

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

  return <DataTable title="Proxy nguồn" table={table} />;
}
