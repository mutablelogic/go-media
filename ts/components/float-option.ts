import { LitElement, css, html } from "lit";
import { ifDefined } from "lit/directives/if-defined.js";
import "@carbon/web-components/es/components/slider/index.js";
import "@carbon/web-components/es/components/number-input/index.js";
import "@carbon/web-components/es/components/button/index.js";
import type { Option } from "../schema/option.js";

// Above this range size, a slider's drag precision becomes unusable. ffmpeg
// defaults unbounded float options to FLT_MIN/FLT_MAX (~±3.4e38), so most
// options technically have min/max set even though they were never meant to
// be sliders — range size, not mere presence of min/max, decides the control.
const MAX_SLIDER_RANGE = 1000;

// Number of steps a slider's drag should be divided into, when its range
// isn't made of whole integers.
const SLIDER_STEPS = 100;

/**
 * Editor for a float/double-typed codec/format Option with no `const`
 * choices (those go through enum-option instead): a slider when min/max are
 * both present and span a small enough range to drag meaningfully,
 * otherwise a plain number input that accepts arbitrary decimals. The
 * option's default is shown alongside (and restorable), and a disabled
 * state is available for read-only contexts.
 */
export class FloatOption extends LitElement {
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

  private get defaultValue(): number {
    return typeof this.option.default === "number" ? this.option.default : 0;
  }

  private get min(): number | undefined {
    return typeof this.option.min === "number" ? this.option.min : undefined;
  }

  private get max(): number | undefined {
    return typeof this.option.max === "number" ? this.option.max : undefined;
  }

  private get useSlider(): boolean {
    return this.min !== undefined && this.max !== undefined && this.max - this.min <= MAX_SLIDER_RANGE;
  }

  // A step fine enough to divide the range into SLIDER_STEPS increments.
  private get sliderStep(): number {
    if (this.min === undefined || this.max === undefined) return 1;
    return (this.max - this.min) / SLIDER_STEPS || 1;
  }

  // Seed the current value from the option's default on first render only,
  // so callers don't have to duplicate `option.default` into a `value`
  // binding — but a caller that does bind `.value` explicitly still wins.
  protected willUpdate(changed: Map<string, unknown>) {
    if (changed.has("option") && !this.hasUpdated) {
      this.value = this.defaultValue;
    }
  }

  private handleSliderChange(event: CustomEvent<{ value: number }>) {
    this.value = event.detail.value;
  }

  private handleNumberChange(event: CustomEvent<{ value: string | number }>) {
    const parsed = Number(event.detail.value);
    if (!Number.isNaN(parsed)) {
      this.value = parsed;
    }
  }

  private resetToDefault() {
    this.value = this.defaultValue;
  }

  render() {
    const isDefault = this.value === this.defaultValue;

    return html`
      ${this.option.description ? html`<p class="description">${this.option.description}</p>` : ""}
      ${this.useSlider
        ? html`
            <cds-slider
              label-text=${this.option.name ?? "Option"}
              min=${this.min ?? 0}
              max=${this.max ?? 0}
              step=${this.sliderStep}
              value=${this.value}
              ?disabled=${this.disabled}
              @cds-slider-changed=${this.handleSliderChange}
            ></cds-slider>
          `
        : html`
            <cds-number-input
              label=${this.option.name ?? "Option"}
              type="number"
              step="any"
              min=${ifDefined(this.min)}
              max=${ifDefined(this.max)}
              value=${this.value}
              ?disabled=${this.disabled}
              @cds-number-input=${this.handleNumberChange}
            ></cds-number-input>
          `}
      <div class="default-row">
        <span>Default: ${this.defaultValue}</span>
        ${!isDefault
          ? html`<cds-button kind="ghost" size="sm" @click=${this.resetToDefault}>Reset to default</cds-button>`
          : ""}
      </div>
    `;
  }
}

customElements.define("float-option", FloatOption);
