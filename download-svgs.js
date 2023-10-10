import { spawnSync } from "bun";
import { join } from "path";
import { mkdirSync, readFileSync, rmdirSync, writeFileSync } from "fs";
import { locate } from "@iconify/json";
import { IconSet } from "@iconify/tools";
import { cleanupSVG } from "@iconify/tools/lib/svg/cleanup";
import { runSVGO } from "@iconify/tools/lib/optimise/svgo";

const outputDir = "svg";

async function downloadSVGs(sets) {
  for (const set of sets) {
    const targetDir = join(outputDir, set);
    mkdirSync(targetDir, { recursive: true });
    const filename = locate(set);
    const data = JSON.parse(readFileSync(filename, "utf8"));
    const iconSet = new IconSet(data);
    iconSet.forEach((name) => {
      const svg = iconSet.toSVG(name);
      if (!svg) {
        console.log("Missing ${name}");
        return;
      }
      cleanupSVG(svg);
      runSVGO(svg);
      if (svg.viewBox.width != svg.viewBox.height) {
        const max = Math.max(svg.viewBox.width, svg.viewBox.height);
        svg.viewBox.width = max;
        svg.viewBox.height = max;
      }

      let svgString = svg.toString({ width: 48, height: 48 });
      if (set == "arcticons") {
        svgString = svgString.replaceAll(
          'stroke="currentColor"',
          'stroke-width="3px" stroke="currentColor"'
        );
      }
      const basename = `${name}.svg`;
      writeFileSync(join(targetDir, basename), svgString, "utf8");
    });
  }
}

async function main() {
  rmdirSync(outputDir, { force: true, recursive: true });
  mkdirSync(outputDir, { recursive: true });
  console.log("downloading arcticons");
  downloadSVGs(["arcticons"]);
  try {
    spawnSync(["npx", "oslllo-svg-fixer", "-s", "svg/arcticons", "-d", "svg/arcticons"]);
  } catch (e) {
    // ignore
  }
  console.log("downloading others");
  downloadSVGs(["fa6-solid", "fa6-regular", "fa6-brands", "mdi", "fluent-emoji-high-contrast"]);
}

main();
