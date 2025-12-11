import {
  AppState,
  listenerServerManager,
  ManagedLocalListener,
  ManagedProxyServer,
} from "@bindings/go-proxy/models";
import { GetAppState, GetManager } from "@bindings/go-proxy/myservice";
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
  servers: ManagedProxyServer[];
  listeners: ManagedLocalListener[];
  fetchManager: () => Promise<void>;
}>(
  (set) => ({
    manager: null,
    servers: [],
    listeners: [],
    fetchManager: () =>
      GetManager().then((manager) => {
        if (manager) {
          set({
            manager,
            servers: Object.values(manager.Servers ?? {}).filter(Boolean),
            listeners: Object.values(manager.Listeners ?? {}).filter(Boolean),
          });
        }
      }),
  }),
  equalJson,
);

export const useAppStateStore = create<{
  state: AppState | null;
  fetchState: () => Promise<void>;
}>(
  (set) => ({
    state: null,
    fetchState: () => GetAppState().then((state) => set({ state })),
  }),
  equalJson,
);

/**
 * @deprecated This hook only exists for initial release
 */
export const findMatchingListener = (
  serverId: string,
  listeners: ManagedLocalListener[],
) => {
  for (const listener of Object.values(listeners ?? {})) {
    if (serverId in (listener?.Listener?.Filter.ServerIds ?? {})) {
      return listener;
    }
  }
};
