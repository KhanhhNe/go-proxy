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

export function pick<
  Inp extends object,
  Keys extends keyof Inp,
  Res extends Pick<Inp, Keys>,
>(...keys: Keys[]): (obj: Inp) => Res {
  return (obj: Inp) => {
    const res = {} as Res;
    for (const k of keys) {
      if (k in obj) {
        // @ts-expect-error Same keys used here
        res[k] = obj[k];
      }
    }
    return res;
  };
}

export function equalJson(a: any, b: any) {
  return JSON.stringify(a) === JSON.stringify(b);
}
