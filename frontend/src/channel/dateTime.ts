import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";

dayjs.extend(customParseFormat);

const dateTimeFormats = [
  "YYYY-MM-DD HH:mm:ss",
  "YYYY-MM-DDTHH:mm:ss",
  "YYYY/MM/DD HH:mm:ss",
  "YYYY-MM-DD",
  "YYYY/MM/DD",
  "YYYYMMDDHHmmss",
  "YYYYMMDD",
  "YYYY-MM",
  "MM/YY",
  "MM/YYYY",
];

export function formatDateTime(value: string | number | Date | null | undefined) {
  if (value === null || value === undefined || value === "") {
    return "—";
  }

  const normalized = typeof value === "string" ? value.trim() : value;
  if (normalized === "") {
    return "—";
  }

  let parsed;
  if (typeof normalized === "number") {
    parsed = Math.abs(normalized) < 100_000_000_000
      ? dayjs.unix(normalized)
      : dayjs(normalized);
  } else if (typeof normalized === "string" && /^-?\d{10}(\.\d+)?$/.test(normalized)) {
    parsed = dayjs.unix(Number(normalized));
  } else if (typeof normalized === "string" && /^-?\d{13}$/.test(normalized)) {
    parsed = dayjs(Number(normalized));
  } else if (typeof normalized === "string") {
    parsed = dayjs(normalized, dateTimeFormats, true);
    if (!parsed.isValid()) {
      parsed = dayjs(normalized);
    }
  } else {
    parsed = dayjs(normalized);
  }

  return parsed.isValid() ? parsed.format("YYYY-MM-DD HH:mm:ss") : "—";
}
