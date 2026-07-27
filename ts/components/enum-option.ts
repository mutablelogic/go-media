import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/radio-button/index.js";
import "@carbon/web-components/es/components/button/index.js";
import type { Option, OptionValue } from "../schema/option.js";

/**
 * Editor for an Option with a `const` list of named choices, where the
 * underlying type (int, float, string, ...) is a single mutually-exclusive
 * value rather than a bitmask — unlike flags-option, these render as a
 * radio group. Some choices (notably sample_format/sample_rate/
 * channel_layout, built by hand in codec.go rather than from ffmpeg's own
 * AVOption consts) have no `name`, so those fall back to showing their
 * value as the label.
 */
export class EnumOption extends LitElement {
  static properties = {
    option: { attribute: false },
    value: { attribute: false },
    disabled: { type: Boolean, reflect: true },
  };

  declare option: Option;
  declare value: OptionValue | undefined;
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
    this.value = undefined;
    this.disabled = false;
  }

  private get choices(): Option[] {
    return this.option.const ?? [];
  }

  private labelFor(choice: Option): string {
    return choice.name || String(choice.default);
  }

  // Seed the current value from the option's default on first render only,
  // so callers don't have to duplicate `option.default` into a `value`
  // binding — but a caller that does bind `.value` explicitly still wins.
  protected willUpdate(changed: Map<string, unknown>) {
    if (changed.has("option") && !this.hasUpdated) {
      this.value = this.option.default;
    }
  }

  private handleChange(event: CustomEvent<{ value: string }>) {
    const choice = this.choices.find((c) => String(c.default) === event.detail.value);
    if (choice) {
      this.value = choice.default;
    }
  }

  private resetToDefault() {
    this.value = this.option.default;
  }

  render() {
    const isDefault = this.value === this.option.default;
    const defaultChoice = this.choices.find((c) => c.default === this.option.default);

    return html`
      ${this.option.description ? html`<p class="description">${this.option.description}</p>` : ""}
      <cds-radio-button-group
        legend-text=${this.option.name ?? "Option"}
        orientation="vertical"
        value=${String(this.value)}
        ?disabled=${this.disabled}
        @cds-radio-button-group-changed=${this.handleChange}
      >
        ${this.choices.map(
          (choice) => html`
            <cds-radio-button label-text=${this.labelFor(choice)} value=${String(choice.default)}></cds-radio-button>
          `,
        )}
      </cds-radio-button-group>
      <div class="default-row">
        <span>Default: ${defaultChoice ? this.labelFor(defaultChoice) : String(this.option.default ?? "None")}</span>
        ${!isDefault
          ? html`<cds-button kind="ghost" size="sm" @click=${this.resetToDefault}>Reset to default</cds-button>`
          : ""}
      </div>
    `;
  }
}

customElements.define("enum-option", EnumOption);
