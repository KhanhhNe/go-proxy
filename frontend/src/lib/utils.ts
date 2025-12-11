import { ManagedProxyServer } from "@bindings/go-proxy";
import { clsx, type ClassValue } from "clsx";
import { useEffect, useState } from "react";
import { twMerge } from "tailwind-merge";
import countriesJson from "./countries.json";

export const countries = countriesJson;
export const COUNTRY_CODES = countries.map((c) => c.code);
export const PROTOCOLS = {
  SOCKS5: "socks5",
  HTTP: "http",
  SSH: "ssh",
};

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

export function sortBy<T>(arr: T[], ...funcs: ((a: T) => any)[]) {
  const res = [...arr];

  res.sort((a, b) => {
    for (const f of funcs) {
      if (f(a) > f(b)) {
        return 1;
      }
    }

    return Number(a > b);
  });

  return res;
}

export function getTags(tags: Record<string, boolean>) {
  return sortBy(
    Object.entries(tags)
      .filter(([, val]) => val)
      .map(([key]) => key),
    (t) => COUNTRY_CODES.includes(t),
    (t) => Object.values(PROTOCOLS).includes(t),
  );
}

export function getServerString(server: ManagedProxyServer["Server"]) {
  if (!server) {
    return "";
  }

  let result = `${server.Host}:${server.Port}`;

  if (server.Auth) {
    result = `${server.Auth.Username}:${server.Auth.Password}@${result}`;
  }

  switch (true) {
    case server.Protocols[PROTOCOLS.HTTP]:
      result = `${PROTOCOLS.HTTP}://${result}`;
      break;
    case server.Protocols[PROTOCOLS.SOCKS5]:
      result = `${PROTOCOLS.SOCKS5}://${result}`;
      break;
  }

  return result;
}

export function useNow() {
  const [now, setNow] = useState(new Date());

  useEffect(() => {
    const int = setInterval(() => setNow(new Date()), 500);

    return () => clearInterval(int);
  }, []);

  return now;
}

export function durationToMs(duration?: number) {
  return duration != null ? Math.round(duration / 1_000_000) : duration;
}

export function debounce<F extends (...args: any[]) => any>(
  func: F,
  waitFor: number,
) {
  let timeout: ReturnType<typeof setTimeout>;

  return (...args: Parameters<F>): void => {
    if (timeout) {
      clearTimeout(timeout);
    }
    timeout = setTimeout(() => {
      func(...args);
    }, waitFor);
  };
}
