import { LitElement, css, html } from "lit";
import "@carbon/web-components/es/components/checkbox/index.js";
import "@carbon/web-components/es/components/button/index.js";
import type { Option } from "../schema/option.js";

/**
 * Editor for a flags-typed codec/format Option (Option.type === "flags"):
 * ffmpeg exposes these as a bitmask, so each `const` choice is a checkbox —
 * checking one ORs its bit into the current value, unchecking clears it via
 * AND-NOT. The option's default is shown alongside (and restorable), and a
 * disabled state is available for read-only contexts.
 */
export class FlagsOption extends LitElement {
  static properties = {
    option: { attribute: false },
    value: { type: Number },
    disabled: { type: Boolean, reflect: true },
  };

  declare option: Option;
  declare value: number;
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
    legend {
      font-size: 0.875rem;
      font-weight: 600;
      color: var(--cds-text-primary, #161616);
      margin-bottom: 0.5rem;
    }
    cds-checkbox {
      display: block;
      margin-bottom: 0.5rem;
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
    this.value = 0;
    this.disabled = false;
  }

  private get choices(): Option[] {
    return this.option.const ?? [];
  }

  private get defaultValue(): number {
    return typeof this.option.default === "number" ? this.option.default : 0;
  }

  // A choice with a zero bit can never be meaningfully toggled via OR/AND-NOT.
  private bitOf(choice: Option): number | undefined {
    return typeof choice.default === "number" && choice.default !== 0 ? choice.default : undefined;
  }

  private namesForValue(value: number): string {
    const names = this.choices
      .filter((c) => {
        const bit = this.bitOf(c);
        return bit !== undefined && (value & bit) === bit;
      })
      .map((c) => c.name);
    return names.length ? names.join(", ") : "None";
  }

  // Seed the current value from the option's default on first render only,
  // so callers don't have to duplicate `option.default` into a `value`
  // binding — but a caller that does bind `.value` explicitly still wins.
  protected willUpdate(changed: Map<string, unknown>) {
    if (changed.has("option") && !this.hasUpdated) {
      this.value = this.defaultValue;
    }
  }

  private handleToggle(bit: number, event: CustomEvent<{ checked: boolean }>) {
    this.value = event.detail.checked ? this.value | bit : this.value & ~bit;
  }

  private resetToDefault() {
    this.value = this.defaultValue;
  }

  render() {
    const isDefault = this.value === this.defaultValue;

    return html`
      ${this.option.description ? html`<p class="description">${this.option.description}</p>` : ""}
      <fieldset ?disabled=${this.disabled}>
        <legend>${this.option.name ?? "Option"}</legend>
        ${this.choices.map((choice) => {
          const bit = this.bitOf(choice);
          if (bit === undefined) return "";
          return html`
            <cds-checkbox
              label-text=${choice.name ?? ""}
              ?checked=${(this.value & bit) === bit}
              ?disabled=${this.disabled}
              @cds-checkbox-changed=${(e: CustomEvent<{ checked: boolean }>) => this.handleToggle(bit, e)}
            ></cds-checkbox>
          `;
        })}
      </fieldset>
      <div class="default-row">
        <span>Default: ${this.namesForValue(this.defaultValue)}</span>
        ${!isDefault
          ? html`<cds-button kind="ghost" size="sm" @click=${this.resetToDefault}>Reset to default</cds-button>`
          : ""}
      </div>
    `;
  }
}

customElements.define("flags-option", FlagsOption);
