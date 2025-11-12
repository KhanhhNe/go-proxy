import { create } from "zustand";
import type {} from "@redux-devtools/extension"; // Required for zustand IDE typing

export const PAGES = {
  index: "index",
  servers: "servers",
} as const;
type PageName = (typeof PAGES)[keyof typeof PAGES];

export const usePageStore = create<{
  page: PageName;
  setPage: (p: PageName) => void;
}>((set) => ({
  page: PAGES.index,
  setPage: (page: PageName) => set({ page }),
}));
