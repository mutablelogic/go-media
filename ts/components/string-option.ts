import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/text-input/index.js";
import "@carbon/web-components/es/components/button/index.js";
import type { Option } from "../schema/option.js";

/**
 * Editor for a string-typed codec/format Option with no `const` choices
 * (those go through enum-option instead): a plain text input. Many of these
 * (e.g. ffmpeg's rc_eq expression options) have no default at all, in which
 * case the default row reads "None" rather than showing an empty value. A
 * disabled state is available for read-only contexts.
 */
export class StringOption extends LitElement {
  static properties = {
    option: { attribute: false },
    value: { type: String },
    disabled: { type: Boolean, reflect: true },
  };

  declare option: Option;
  declare value: string;
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
    this.value = "";
    this.disabled = false;
  }

  private get defaultValue(): string {
    return typeof this.option.default === "string" ? this.option.default : "";
  }

  // Seed the current value from the option's default on first render only,
  // so callers don't have to duplicate `option.default` into a `value`
  // binding — but a caller that does bind `.value` explicitly still wins.
  protected willUpdate(changed: Map<string, unknown>) {
    if (changed.has("option") && !this.hasUpdated) {
      this.value = this.defaultValue;
    }
  }

  private handleInput(event: Event) {
    this.value = (event.target as HTMLInputElement).value;
  }

  private resetToDefault() {
    this.value = this.defaultValue;
  }

  render() {
    const isDefault = this.value === this.defaultValue;

    return html`
      ${this.option.description ? html`<p class="description">${this.option.description}</p>` : ""}
      <cds-text-input
        label=${this.option.name ?? "Option"}
        value=${this.value}
        ?disabled=${this.disabled}
        @input=${this.handleInput}
      ></cds-text-input>
      <div class="default-row">
        <span>Default: ${this.option.default !== undefined ? this.defaultValue : "None"}</span>
        ${!isDefault
          ? html`<cds-button kind="ghost" size="sm" @click=${this.resetToDefault}>Reset to default</cds-button>`
          : ""}
      </div>
    `;
  }
}

customElements.define("string-option", StringOption);
