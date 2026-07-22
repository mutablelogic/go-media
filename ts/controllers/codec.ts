import type { ReactiveController, ReactiveControllerHost } from "lit";
import { Client } from "../httpclient/client.js";
import { listCodecs } from "../httpclient/codec.js";
import type { Codec, CodecFilterType } from "../schema/codec.js";

const DEFAULT_PAGE_SIZE = 20;

/**
 * Fetches a page of codecs when its host connects or the filter/page changes,
 * and calls `host.requestUpdate()` once the data (or an error) arrives so the
 * host component redisplays with the fetched codecs.
 */
export class CodecController implements ReactiveController {
  private readonly host: ReactiveControllerHost;
  private readonly client: Client;

  codecs: Codec[] = [];
  total = 0;
  loading = false;
  error?: string;

  type?: CodecFilterType;
  page = 1;
  pageSize = DEFAULT_PAGE_SIZE;

  constructor(host: ReactiveControllerHost, endpoint = "http://127.0.0.1:8084/api") {
    this.host = host;
    this.client = new Client(endpoint);
    host.addController(this);
  }

  hostConnected(): void {
    this.refresh();
  }

  /** Filter by codec type (or clear the filter with undefined) and jump back to page 1. */
  setType(type: CodecFilterType | undefined): void {
    this.type = type;
    this.page = 1;
    this.refresh();
  }

  /** Move to a different page and/or page size. */
  setPage(page: number, pageSize: number = this.pageSize): void {
    this.page = page;
    this.pageSize = pageSize;
    this.refresh();
  }

  async refresh(): Promise<void> {
    this.loading = true;
    this.error = undefined;
    this.host.requestUpdate();

    try {
      const list = await listCodecs(this.client, {
        type: this.type,
        offset: (this.page - 1) * this.pageSize,
        limit: this.pageSize,
      });
      this.codecs = list.body ?? [];
      this.total = list.count;
    } catch (err) {
      this.error = err instanceof Error ? err.message : String(err);
    } finally {
      this.loading = false;
      this.host.requestUpdate();
    }
  }
}
