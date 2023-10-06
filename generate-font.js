import { writeFileSync } from "fs";
import { createSVG, createTTF, createWOFF2 } from "svgtofont/lib/utils.js";

const infoData = {};

const options = {
  src: "svg",
  dist: "fonts",
  emptyDist: true,
  fontName: "paisa",
  svg2ttf: {},
  svgicons2svgfont: {
    normalize: true,
    fontWeight: 900,
    fontHeight: 1000,
    fontStyle: "normal",
    fixedWidth: true,
    centerHorizontally: true,
    centerVertically: true
  },
  useNameAsUnicode: false,
  symbolNameDelimiter: "-",
  css: false,
  getIconUnicode: (name, _unicode, startUnicode) => {
    startUnicode++;
    const [library, symbol] = name.split(":");
    if (!infoData[library]) {
      infoData[library] = {};
    }

    infoData[library][symbol] = startUnicode;
    return [String.fromCodePoint(startUnicode), startUnicode];
  },
  startUnicode: 0xf0000,
  outSVGReact: false,
  generateInfoData: true
};

async function createFont() {
  await createSVG(options);
  const ttf = await createTTF(options);
  await createWOFF2(options, ttf);

  if (options.generateInfoData) {
    writeFileSync("fonts/info.json", JSON.stringify(infoData), "utf8");
  }
}

createFont();
