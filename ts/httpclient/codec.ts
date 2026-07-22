// Mirrors profile/httpclient/codec.go

import type { Client } from "./client.js";
import type { Codec, CodecFilterType, CodecList, CodecListRequest } from "../schema/codec.js";

// The server's query-string binder (httprequest.Query) reads "type" as a raw
// integer into CodecType's underlying int, rather than through the string-based
// JSON (un)marshaling CodecType otherwise uses. These ordinals are libavutil's
// AVMediaType C enum values, which are stable across ffmpeg versions.
const AVMEDIA_TYPE: Record<CodecFilterType, number> = {
  video: 0,
  audio: 1,
  subtitle: 3,
};

function query(req: CodecListRequest): URLSearchParams {
  const params = new URLSearchParams();
  if (req.type) {
    params.set("type", String(AVMEDIA_TYPE[req.type]));
  }
  if (req.offset) {
    params.set("offset", String(req.offset));
  }
  if (req.limit !== undefined) {
    params.set("limit", String(req.limit));
  }
  return params;
}

export function listCodecs(client: Client, req: CodecListRequest = {}): Promise<CodecList> {
  return client.request<CodecList>(["codec"], query(req));
}

export function getCodec(client: Client, name: string): Promise<Codec> {
  return client.request<Codec>(["codec", name]);
}
