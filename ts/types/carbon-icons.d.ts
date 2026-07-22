// @carbon/icons ships each icon as a bare .js module with no type declarations.
declare module "@carbon/icons/es/*" {
  import type IconDescriptor from "@carbon/icon-helpers/es/types.js";
  const descriptor: IconDescriptor;
  export default descriptor;
}
