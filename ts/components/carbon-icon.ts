import { toSVG } from "@carbon/icon-helpers";
import type IconDescriptor from "@carbon/icon-helpers/es/types.js";

/**
 * Converts an @carbon/icons descriptor into an SVGElement ready to slot into
 * a Carbon button's `icon` slot. Lit accepts an SVGElement directly as a
 * child value, so callers can cache the result and reuse the same node
 * across renders instead of rebuilding it each time.
 *
 * Builds the root <svg> by hand rather than calling toSVG(descriptor)
 * directly: @carbon/icon-helpers@10.47.0's toSVG() computes the merged
 * focusable/aria-hidden/preserveAspectRatio defaults correctly, but then
 * applies values from the *pre-merge* attrs object when setting them, so the
 * browser logs an invalid-attribute-value error the instant it sets
 * focusable="undefined" etc. Child elements (plain <path>s) don't hit that
 * code path, so toSVG is still used for those.
 */
export function icon(descriptor: IconDescriptor): SVGElement {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  for (const [key, value] of Object.entries(descriptor.attrs ?? {})) {
    svg.setAttribute(key, value);
  }
  svg.setAttribute("focusable", "false");
  svg.setAttribute("aria-hidden", "true");
  svg.setAttribute("preserveAspectRatio", "xMidYMid meet");
  svg.setAttribute("slot", "icon");
  for (const child of descriptor.content ?? []) {
    svg.appendChild(toSVG(child));
  }
  return svg;
}
