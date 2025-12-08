import { AppState, listenerServerManager } from "@bindings/go-proxy/models";
import type {} from "@redux-devtools/extension"; // Required for zustand IDE typing
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
  manager: listenerServerManager | null;
  setManager: (m: listenerServerManager) => void;
}>(
  (set) => ({
    manager: null,
    setManager: (manager: listenerServerManager) => set({ manager }),
  }),
  equalJson,
);

export const useAppStateStore = create<{
  state: AppState | null;
  setState: (s: AppState) => void;
}>(
  (set) => ({
    state: null,
    setState: (state: AppState) => set({ state }),
  }),
  equalJson,
);

/**
 * @deprecated This hook only exists for initial release
 */
export const useMatchingListener = (serverId: string) => {
  const manager = useManagerStore((s) => s.manager);

  for (const listener of Object.values(manager?.Listeners ?? {})) {
    if (serverId in (listener?.Listener?.Filter.ServerIds ?? {})) {
      return listener;
    }
  }
};
