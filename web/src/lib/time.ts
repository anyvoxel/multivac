export function formatDateTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString();
}

export function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const yyyy = d.getFullYear();
  const mm = pad2(d.getMonth() + 1);
  const dd = pad2(d.getDate());
  return `${yyyy}/${mm}/${dd}`;
}

function pad2(n: number): string {
  return n < 10 ? `0${n}` : `${n}`;
}

function toDate(input?: Date | string): Date | null {
  if (input === undefined) return null;
  const d = input instanceof Date ? new Date(input.getTime()) : new Date(input);
  if (Number.isNaN(d.getTime())) return null;
  return d;
}

export function startOfLocalDay(input?: Date | string): Date {
  const base = toDate(input) ?? new Date();
  return new Date(base.getFullYear(), base.getMonth(), base.getDate());
}

export function addDays(input: Date | string, days: number): Date {
  const base = startOfLocalDay(input);
  return new Date(base.getFullYear(), base.getMonth(), base.getDate() + days);
}

export function toLocalDateKey(input: Date | string): string {
  const d = startOfLocalDay(input);
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
}

export function isSameLocalDay(a: Date | string, b: Date | string): boolean {
  return toLocalDateKey(a) === toLocalDateKey(b);
}

export function startOfWeek(input: Date | string, weekStartsOn = 0): Date {
  const d = startOfLocalDay(input);
  const day = d.getDay();
  const offset = (day - weekStartsOn + 7) % 7;
  return addDays(d, -offset);
}

export function getWeekDays(input: Date | string, weekStartsOn = 0): Date[] {
  const start = startOfWeek(input, weekStartsOn);
  return Array.from({ length: 7 }, (_, index) => addDays(start, index));
}

export function formatWeekRange(input: Date | string, weekStartsOn = 0): string {
  const days = getWeekDays(input, weekStartsOn);
  const first = days[0];
  const last = days[6];
  if (first.getFullYear() === last.getFullYear() && first.getMonth() === last.getMonth()) {
    return `${first.getFullYear()}年${first.getMonth() + 1}月${first.getDate()}日 - ${last.getDate()}日`;
  }
  if (first.getFullYear() === last.getFullYear()) {
    return `${first.getFullYear()}年${first.getMonth() + 1}月${first.getDate()}日 - ${last.getMonth() + 1}月${last.getDate()}日`;
  }
  return `${first.getFullYear()}年${first.getMonth() + 1}月${first.getDate()}日 - ${last.getFullYear()}年${last.getMonth() + 1}月${last.getDate()}日`;
}

export function formatWeekdayShort(input: Date | string): string {
  const d = startOfLocalDay(input);
  return ["周日", "周一", "周二", "周三", "周四", "周五", "周六"][d.getDay()];
}

export function formatMonthDay(input: Date | string): string {
  const d = startOfLocalDay(input);
  return `${d.getMonth() + 1}月${d.getDate()}日`;
}

// toDateTimeLocalValue converts ISO time string into input[type=datetime-local] value.
export function toDateTimeLocalValue(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const yyyy = d.getFullYear();
  const mm = pad2(d.getMonth() + 1);
  const dd = pad2(d.getDate());
  const hh = pad2(d.getHours());
  const mi = pad2(d.getMinutes());
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`;
}

// fromDateTimeLocalValue converts datetime-local input value into RFC3339 string.
export function fromDateTimeLocalValue(v: string): string | undefined {
  if (!v) return undefined;
  const d = new Date(v);
  if (Number.isNaN(d.getTime())) return undefined;
  return d.toISOString();
}

export function toDateValue(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const yyyy = d.getFullYear();
  const mm = pad2(d.getMonth() + 1);
  const dd = pad2(d.getDate());
  return `${yyyy}-${mm}-${dd}`;
}

export function fromDateValue(v: string): string | undefined {
  if (!v) return undefined;
  const d = new Date(`${v}T00:00:00`);
  if (Number.isNaN(d.getTime())) return undefined;
  return d.toISOString();
}
