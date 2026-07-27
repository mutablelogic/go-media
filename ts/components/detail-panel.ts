import { LitElement, css, html } from "lit";
import Close20 from "@carbon/icons/es/close/20.js";
import "@carbon/web-components/es/components/icon-button/index.js";
import { icon } from "./carbon-icon.js";

/**
 * A modal panel anchored to the right edge of the viewport, with a dimmed
 * backdrop. Closes on backdrop click, the close button, or Escape.
 *
 * cds-side-panel (Carbon Web Components) was considered for this but ships
 * with no bundled CSS at all — it's deprecated in favor of a component now
 * only available in the separate, pre-1.0 @carbon/ibm-products-web-components
 * package — so this reimplements just the bit we need using Carbon's design
 * tokens directly.
 */
export class DetailPanel extends LitElement {
  static properties = {
    open: { type: Boolean, reflect: true },
    panelTitle: { attribute: "panel-title" },
    panelSubtitle: { attribute: "panel-subtitle" },
  };

  declare open: boolean;
  declare panelTitle: string;
  declare panelSubtitle: string;

  private readonly closeIcon = icon(Close20);

  static styles = css`
    :host {
      display: contents;
    }
    .overlay {
      position: fixed;
      inset: 0;
      z-index: 9000;
      background: var(--cds-overlay, rgba(22, 22, 22, 0.5));
      opacity: 0;
      pointer-events: none;
      transition: opacity 150ms ease;
    }
    :host([open]) .overlay {
      opacity: 1;
      pointer-events: auto;
    }
    .panel {
      position: fixed;
      inset-block: 0;
      inset-inline-end: 0;
      z-index: 9001;
      inline-size: 20rem;
      max-inline-size: 90vw;
      display: flex;
      flex-direction: column;
      background: var(--cds-layer, #f4f4f4);
      box-shadow: -0.5rem 0 1rem rgba(0, 0, 0, 0.15);
      transform: translateX(100%);
      transition: transform 150ms ease;
      visibility: hidden;
    }
    :host([open]) .panel {
      transform: translateX(0);
      visibility: visible;
    }
    header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 0.5rem;
      padding: 1rem 1rem 0.5rem;
      border-block-end: 1px solid var(--cds-border-subtle, #e0e0e0);
    }
    h2 {
      margin: 0;
      font-size: 1rem;
      font-weight: 600;
      color: var(--cds-text-primary, #161616);
    }
    p.subtitle {
      margin: 0.25rem 0 0;
      color: var(--cds-text-secondary, #525252);
      font-size: 0.875rem;
    }
    .body {
      flex: 1;
      overflow-y: auto;
      padding: 1rem;
      color: var(--cds-text-primary, #161616);
    }
  `;

  constructor() {
    super();
    this.open = false;
    this.panelTitle = "";
    this.panelSubtitle = "";
  }

  connectedCallback() {
    super.connectedCallback();
    document.addEventListener("keydown", this.handleKeydown);
  }

  disconnectedCallback() {
    document.removeEventListener("keydown", this.handleKeydown);
    super.disconnectedCallback();
  }

  private handleKeydown = (event: KeyboardEvent) => {
    if (this.open && event.key === "Escape") {
      this.close();
    }
  };

  private close() {
    this.open = false;
    this.dispatchEvent(new Event("close"));
  }

  render() {
    return html`
      <div class="overlay" @click=${this.close}></div>
      <div class="panel" role="complementary" aria-hidden=${!this.open}>
        <header>
          <div>
            <h2>${this.panelTitle}</h2>
            ${this.panelSubtitle ? html`<p class="subtitle">${this.panelSubtitle}</p>` : ""}
          </div>
          <cds-icon-button kind="ghost" @click=${this.close}>
            ${this.closeIcon}
            <span slot="tooltip-content">Close</span>
          </cds-icon-button>
        </header>
        <div class="body">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define("detail-panel", DetailPanel);
