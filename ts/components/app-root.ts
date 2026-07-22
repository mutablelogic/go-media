import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/ui-shell/index.js";
import "./home-view.js";

/**
 * Application shell: Carbon header + side nav wrapping the routed content area.
 */
export class AppRoot extends LitElement {
  static properties = {
    sideNavExpanded: { state: true },
  };

  declare sideNavExpanded: boolean;

  static styles = css`
    :host {
      display: block;
    }
    main {
      padding: 2rem;
      margin-top: 3rem;
    }
    @media (min-width: 66rem) {
      main {
        margin-inline-start: 16rem;
      }
    }
  `;

  constructor() {
    super();
    this.sideNavExpanded = false;
  }

  private handleMenuToggle(event: CustomEvent<{ active: boolean }>) {
    this.sideNavExpanded = event.detail.active;
  }

  render() {
    return html`
      <cds-header aria-label="go-media">
        <cds-header-menu-button
          button-label-active="Close menu"
          button-label-inactive="Open menu"
          ?active=${this.sideNavExpanded}
          @cds-header-menu-button-toggled=${this.handleMenuToggle}
        ></cds-header-menu-button>
        <cds-header-name href="/" prefix="go">media</cds-header-name>
      </cds-header>
      <cds-side-nav
        ?expanded=${this.sideNavExpanded}
        collapse-mode="responsive"
      >
        <cds-side-nav-items>
          <cds-side-nav-link href="/">Home</cds-side-nav-link>
        </cds-side-nav-items>
      </cds-side-nav>
      <main>
        <home-view></home-view>
      </main>
    `;
  }
}

customElements.define("app-root", AppRoot);
