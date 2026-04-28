const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

export function resolveMediaURL(value: string): string {
  if (!value) {
    return value;
  }

  try {
    const parsed = new URL(value);
    if (parsed.pathname !== "/media/object" || !API_BASE) {
      return value;
    }

    const apiURL = new URL(API_BASE);
    parsed.protocol = apiURL.protocol;
    parsed.host = apiURL.host;
    return parsed.toString();
  } catch {
    return value;
  }
}
