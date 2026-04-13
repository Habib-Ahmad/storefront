export type LogisticsAddressFields = {
  streetAddress: string;
  city: string;
  state: string;
  country: string;
};

const defaultCountry = "Nigeria";

function normalize(value: string) {
  return value.trim();
}

export function emptyLogisticsAddress(): LogisticsAddressFields {
  return {
    streetAddress: "",
    city: "",
    state: "",
    country: defaultCountry,
  };
}

export function parseLogisticsAddress(value?: string | null): LogisticsAddressFields {
  if (!value?.trim()) {
    return emptyLogisticsAddress();
  }

  const segments = value
    .split(",")
    .map((segment) => normalize(segment))
    .filter(Boolean);

  if (segments.length >= 4) {
    return {
      streetAddress: segments.slice(0, -3).join(", "),
      city: segments.at(-3) ?? "",
      state: segments.at(-2) ?? "",
      country: segments.at(-1) ?? defaultCountry,
    };
  }

  if (segments.length === 3) {
    return {
      streetAddress: segments[0],
      city: segments[1],
      state: segments[2],
      country: defaultCountry,
    };
  }

  if (segments.length === 2) {
    return {
      streetAddress: segments[0],
      city: segments[1],
      state: "",
      country: defaultCountry,
    };
  }

  return {
    streetAddress: segments[0] ?? "",
    city: "",
    state: "",
    country: defaultCountry,
  };
}

export function formatLogisticsAddress(fields: LogisticsAddressFields): string {
  return [fields.streetAddress, fields.city, fields.state, fields.country]
    .map(normalize)
    .filter(Boolean)
    .join(", ");
}

export function isLogisticsAddressComplete(fields: LogisticsAddressFields): boolean {
  return [fields.streetAddress, fields.city, fields.state, fields.country].every(
    (value) => normalize(value).length > 0,
  );
}

export function missingLogisticsAddressFields(fields: LogisticsAddressFields): string[] {
  const missing: string[] = [];
  if (!normalize(fields.streetAddress)) {
    missing.push("pickup street address");
  }
  if (!normalize(fields.city)) {
    missing.push("pickup city");
  }
  if (!normalize(fields.state)) {
    missing.push("pickup state");
  }
  if (!normalize(fields.country)) {
    missing.push("pickup country");
  }
  return missing;
}
