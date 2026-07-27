// Mirrors profile/schema/codec.go

import type { Option } from "./option.js";

/** Codec media type, as returned by the server. */
export type CodecType = "audio" | "video" | "subtitle" | "data" | "attachment" | "unknown";

/** Codec types accepted when filtering a codec list request. */
export type CodecFilterType = "audio" | "video" | "subtitle";

export interface Codec {
  name: string;
  description?: string;
  type: CodecType;
  opts?: Option[];
}

export interface CodecListRequest {
  type?: CodecFilterType;
  offset?: number;
  limit?: number;
}

export interface CodecList extends CodecListRequest {
  count: number;
  body?: Codec[];
}
