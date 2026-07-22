import { build, context } from "esbuild";
import { cpSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const srcDir = join(here, "assets");
const outDir = join(here, "..", "build", "ts", "dist");
const watch = process.argv.includes("--watch");
const serve = process.argv.includes("--serve");
const port = Number(process.env.PORT) || 8000;

mkdirSync(outDir, { recursive: true });
cpSync(join(srcDir, "index.html"), join(outDir, "index.html"));

const options = {
  entryPoints: [
    { in: join(srcDir, "main.ts"), out: "main" },
    { in: join(srcDir, "global.css"), out: "styles" },
  ],
  bundle: true,
  splitting: false,
  format: "esm",
  target: "es2022",
  sourcemap: true,
  minify: !watch && !serve,
  outdir: outDir,
  logLevel: "info",
};

if (serve) {
  const ctx = await context(options);
  await ctx.watch();
  await ctx.serve({ servedir: outDir, port });
} else if (watch) {
  const ctx = await context(options);
  await ctx.watch();
  console.log("Watching ts/assets for changes...");
} else {
  await build(options);
}
