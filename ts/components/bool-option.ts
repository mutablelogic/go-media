import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/toggle/index.js";
import "@carbon/web-components/es/components/button/index.js";
import type { Option } from "../schema/option.js";

/**
 * Editor for a boolean-typed codec/format Option (Option.type === "bool"):
 * a toggle switch for the current value, with the option's default shown
 * alongside (and restorable) and a disabled state for read-only contexts.
 */
export class BoolOption extends LitElement {
  static properties = {
    option: { attribute: false },
    value: { type: Boolean },
    disabled: { type: Boolean, reflect: true },
  };

  declare option: Option;
  declare value: boolean;
  declare disabled: boolean;

  static styles = css`
    :host {
      display: block;
      margin-block-end: 1rem;
    }
    p.description {
      margin: 0 0 0.5rem;
      color: var(--cds-text-secondary, #525252);
      font-size: 0.875rem;
    }
    .default-row {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: 0.25rem;
    }
    .default-row span {
      color: var(--cds-text-helper, #6f6f6f);
      font-size: 0.75rem;
    }
  `;

  constructor() {
    super();
    this.option = {};
    this.value = false;
    this.disabled = false;
  }

  private get defaultValue(): boolean {
    return this.option.default === true;
  }

  // Seed the current value from the option's default on first render only,
  // so callers don't have to duplicate `option.default` into a `value`
  // binding — but a caller that does bind `.value` explicitly still wins.
  protected willUpdate(changed: Map<string, unknown>) {
    if (changed.has("option") && !this.hasUpdated) {
      this.value = this.defaultValue;
    }
  }

  private handleToggle(event: CustomEvent<{ toggled: boolean }>) {
    this.value = event.detail.toggled;
  }

  private resetToDefault() {
    this.value = this.defaultValue;
  }

  render() {
    const isDefault = this.value === this.defaultValue;

    return html`
      ${this.option.description ? html`<p class="description">${this.option.description}</p>` : ""}
      <cds-toggle
        ?toggled=${this.value}
        ?disabled=${this.disabled}
        @cds-toggle-changed=${this.handleToggle}
      >
        <span slot="label-text">${this.option.name ?? "Option"}</span>
      </cds-toggle>
      <div class="default-row">
        <span>Default: ${this.defaultValue ? "On" : "Off"}</span>
        ${!isDefault
          ? html`<cds-button kind="ghost" size="sm" @click=${this.resetToDefault}>Reset to default</cds-button>`
          : ""}
      </div>
    `;
  }
}

customElements.define("bool-option", BoolOption);
