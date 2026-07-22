// Mirrors profile/httpclient/httpclient.go

/** Error response body the server returns for non-2xx responses (httpresponse.ErrResponse). */
export interface ErrResponse {
  object: string;
  code: number;
  reason?: string;
  detail?: unknown;
}

/** Thrown when a request to the profile API fails. */
export class ClientError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly detail?: unknown,
  ) {
    super(message);
    this.name = "ClientError";
  }
}

/**
 * HTTP client for the profile API.
 *
 * `endpoint` should point at the profile API root, e.g. "http://localhost:8080/api".
 */
export class Client {
  private readonly endpoint: URL;

  constructor(endpoint: string) {
    this.endpoint = new URL(endpoint);
  }

  /** Issue a GET request against endpoint/...path, decoding the JSON response as T. */
  async request<T>(path: string[], query?: URLSearchParams): Promise<T> {
    const url = new URL(this.endpoint);
    url.pathname = [url.pathname.replace(/\/+$/, ""), ...path.map(encodeURIComponent)].join("/");
    if (query) {
      url.search = query.toString();
    }

    const response = await fetch(url, {
      method: "GET",
      headers: { Accept: "application/json" },
    });
    if (!response.ok) {
      throw await errorFromResponse(response);
    }

    return (await response.json()) as T;
  }
}

async function errorFromResponse(response: Response): Promise<ClientError> {
  const text = await response.text();
  if (!text) {
    return new ClientError(response.status, response.statusText || `HTTP ${response.status}`);
  }
  try {
    const body = JSON.parse(text) as ErrResponse;
    return new ClientError(response.status, body.reason || response.statusText, body.detail);
  } catch {
    return new ClientError(response.status, text);
  }
}
