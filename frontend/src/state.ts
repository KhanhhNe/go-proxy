import type {} from "@redux-devtools/extension"; // Required for zustand IDE typing
import { main } from "@wailsjs/go/models";
import { createWithEqualityFn as create } from "zustand/traditional";
import { equalJson } from "./lib/utils";

export const PAGES = {
  index: "index",
  servers: "servers",
} as const;
type PageName = (typeof PAGES)[keyof typeof PAGES];

export const usePageStore = create<{
  page: PageName;
  setPage: (p: PageName) => void;
}>((set) => ({
  page: PAGES.servers,
  setPage: (page: PageName) => set({ page }),
}));

export const useManagerStore = create<{
  manager: main.listenerServerManager | null;
  setManager: (m: main.listenerServerManager) => void;
}>(
  (set) => ({
    manager: null,
    setManager: (manager: main.listenerServerManager) => set({ manager }),
  }),
  equalJson,
);
