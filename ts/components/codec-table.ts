import { LitElement, css, html } from "lit";
import Music16 from "@carbon/icons/es/music/16.js";
import Video16 from "@carbon/icons/es/video/16.js";
import ClosedCaption16 from "@carbon/icons/es/closed-caption/16.js";
import "@carbon/web-components/es/components/data-table/index.js";
import "@carbon/web-components/es/components/icon-button/index.js";
import "@carbon/web-components/es/components/pagination/index.js";
import "@carbon/web-components/es/components/select/index.js";
import { icon } from "./carbon-icon.js";
import "./detail-panel.js";
import "./bool-option.js";
import "./flags-option.js";
import "./enum-option.js";
import "./int-option.js";
import "./float-option.js";
import "./string-option.js";
import { CodecController } from "../controllers/codec.js";
import type { Codec, CodecFilterType } from "../schema/codec.js";

const FILTERS: { label: string; type: CodecFilterType; icon: SVGElement }[] = [
  { label: "Audio", type: "audio", icon: icon(Music16) },
  { label: "Video", type: "video", icon: icon(Video16) },
  { label: "Subtitle", type: "subtitle", icon: icon(ClosedCaption16) },
];

const PAGE_SIZES = [10, 20, 50];

/**
 * Table view of the codecs available on the server, populated by CodecController.
 */
export class CodecTable extends LitElement {
  static properties = {
    selectedCodec: { state: true },
  };

  static styles = css`
    :host {
      display: block;
    }
    cds-table-row {
      cursor: pointer;
    }
  `;

  declare selectedCodec: Codec | null;

  private readonly codecs = new CodecController(this);

  constructor() {
    super();
    this.selectedCodec = null;
  }

  // Clicking the active filter again clears it back to "all".
  private toggleType(type: CodecFilterType) {
    this.codecs.setType(this.codecs.type === type ? undefined : type);
  }

  private handlePageChanged(event: CustomEvent<{ page: number; pageSize: number }>) {
    this.codecs.setPage(event.detail.page, event.detail.pageSize);
  }

  private handleClosePanel() {
    this.selectedCodec = null;
  }

  render() {
    if (this.codecs.error) {
      return html`<p>Failed to load codecs: ${this.codecs.error}</p>`;
    }

    return html`
      <cds-table-toolbar>
        <cds-table-toolbar-content>
          ${FILTERS.map(
            (f) => html`
              <cds-icon-button
                kind=${this.codecs.type === f.type ? "primary" : "ghost"}
                @click=${() => this.toggleType(f.type)}
              >
                ${f.icon}
                <span slot="tooltip-content">${f.label}</span>
              </cds-icon-button>
            `,
          )}
        </cds-table-toolbar-content>
      </cds-table-toolbar>

      ${this.codecs.loading && this.codecs.codecs.length === 0
        ? html`<cds-table-skeleton></cds-table-skeleton>`
        : html`
            <cds-table>
              <cds-table-head>
                <cds-table-header-row>
                  <cds-table-header-cell>Name</cds-table-header-cell>
                  <cds-table-header-cell>Type</cds-table-header-cell>
                  <cds-table-header-cell>Description</cds-table-header-cell>
                </cds-table-header-row>
              </cds-table-head>
              <cds-table-body>
                ${this.codecs.codecs.map(
                  (codec) => html`
                    <cds-table-row @click=${() => (this.selectedCodec = codec)}>
                      <cds-table-cell>${codec.name}</cds-table-cell>
                      <cds-table-cell>${codec.type}</cds-table-cell>
                      <cds-table-cell>${codec.description ?? ""}</cds-table-cell>
                    </cds-table-row>
                  `,
                )}
              </cds-table-body>
            </cds-table>
          `}

      <cds-pagination
        page=${this.codecs.page}
        page-size=${this.codecs.pageSize}
        total-items=${this.codecs.total}
        items-per-page-text="Items per page:"
        @cds-pagination-changed-current=${this.handlePageChanged}
        @cds-page-sizes-select-changed=${this.handlePageChanged}
      >
        ${PAGE_SIZES.map((size) => html`<cds-select-item value=${size}>${size}</cds-select-item>`)}
      </cds-pagination>

      <detail-panel
        ?open=${this.selectedCodec !== null}
        panel-title=${this.selectedCodec?.name ?? ""}
        panel-subtitle=${this.selectedCodec?.type ?? ""}
        @close=${this.handleClosePanel}
      >
        ${this.selectedCodec
          ? html`
              <p>${this.selectedCodec.description ?? "No description available."}</p>
              ${this.selectedCodec.opts?.length
                ? html`
                    <h3>Options</h3>
                    ${this.selectedCodec.opts.map((opt) => {
                      if (opt.type === "bool") {
                        return html`<bool-option .option=${opt}></bool-option>`;
                      }
                      if (opt.type === "flags") {
                        return html`<flags-option .option=${opt}></flags-option>`;
                      }
                      if (opt.const?.length) {
                        return html`<enum-option .option=${opt}></enum-option>`;
                      }
                      if (opt.type === "int" || opt.type === "int64") {
                        return html`<int-option .option=${opt}></int-option>`;
                      }
                      if (opt.type === "float" || opt.type === "double") {
                        return html`<float-option .option=${opt}></float-option>`;
                      }
                      if (opt.type === "string") {
                        return html`<string-option .option=${opt}></string-option>`;
                      }
                      return html`
                        <p>
                          <strong>${opt.name}</strong>${opt.type ? html` (${opt.type})` : ""}
                          ${opt.default !== undefined ? html` — default: ${opt.default}` : ""}
                        </p>
                      `;
                    })}
                  `
                : ""}
            `
          : ""}
      </detail-panel>
    `;
  }
}

customElements.define("codec-table", CodecTable);
