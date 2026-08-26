#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "..");

function walk(dir, out = []) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, ent.name);
    if (ent.isDirectory()) walk(p, out);
    else if (ent.name === "api.ts") out.push(p);
  }
  return out;
}

for (const app of ["admin_web", "super_admin_web"]) {
  const base = path.join(root, "apps", app, "src", "features");
  if (!fs.existsSync(base)) continue;
  for (const file of walk(base)) {
    if (file.includes(`${path.sep}dashboard${path.sep}`)) continue;
    let src = fs.readFileSync(file, "utf8");
    const before = src;
    if (!src.includes("ALLOW_MOCK_FALLBACK")) {
      src = src.replace(
        /^(import .+ from "@\/shared\/api\/client";)/m,
        'import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";\n$1',
      );
    }
    src = src.replace(/\} catch \{/g, "} catch (err) {\n    if (!ALLOW_MOCK_FALLBACK) throw err;");
    src = src.replace(/\} catch \(err\) \{\s*if \(err instanceof ApiError/g, "} catch (err) {\n    if (!ALLOW_MOCK_FALLBACK) throw err;\n    if (err instanceof ApiError");
    if (src !== before) {
      fs.writeFileSync(file, src);
      console.log("guarded", path.relative(root, file));
    }
  }
}
