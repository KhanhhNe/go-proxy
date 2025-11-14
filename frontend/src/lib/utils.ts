import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const BYTE_UNITS = ["B", "KB", "MB", "GB"];
export function formatByte(n: number) {
  let i = 0;

  while (i < BYTE_UNITS.length - 1 && n > 1024) {
    n /= 1024;
    i += 1;
  }

  const fixed = n === Math.round(n) ? Math.round(n).toString() : n.toFixed(1);

  return fixed + " " + BYTE_UNITS[i];
}
