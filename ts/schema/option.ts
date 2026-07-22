// Mirrors profile/schema/option.go

/** Values carried by Option.default/min/max: numbers, strings (including "num/den" rationals), or booleans. */
export type OptionValue = string | number | boolean;

/** A single configuration option for a codec or format, and its constraints. */
export interface Option {
  name?: string;
  description?: string;
  type?: string;
  default?: OptionValue;
  const?: Option[];
  min?: OptionValue;
  max?: OptionValue;
  unit?: string;
}
