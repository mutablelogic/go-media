import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/tile/index.js";
import "@carbon/web-components/es/components/button/index.js";
import "./codec-table.js";

/**
 * Placeholder landing view — replace with real routed content.
 */
export class HomeView extends LitElement {
  static styles = css`
    :host {
      display: block;
      max-width: 48rem;
    }
    cds-tile {
      margin-bottom: 1rem;
    }
    codec-table {
      margin-top: 2rem;
    }
  `;

  render() {
    return html`
      <cds-tile><h1>Codecs</h1></cds-tile>
      <codec-table></codec-table>
    `;
  }
}

customElements.define("home-view", HomeView);
