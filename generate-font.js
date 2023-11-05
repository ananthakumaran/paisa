import _ from "lodash";
import { writeFileSync } from "fs";
import { join } from "path";
import { createSVG, createTTF, createWOFF2 } from "svgtofont/lib/utils.js";

let infoData = {};

let lastUnicode = 0xf0000;

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
    lastUnicode = startUnicode;
    infoData[name] = startUnicode;
    return [String.fromCodePoint(startUnicode), startUnicode];
  },
  startUnicode: 0xf0000,
  outSVGReact: false,
  generateInfoData: true
};

async function createFont(font) {
  infoData = {};
  options.src = join("svg", font);
  options.fontName = font;
  options.startUnicode = lastUnicode++;
  await createSVG(options);
  const ttf = await createTTF(options);
  await createWOFF2(options, ttf);

  if (options.generateInfoData) {
    writeFileSync(`fonts/${font}-info.json`, JSON.stringify({ codepoints: infoData }), "utf8");

    const min = _.min(Object.values(infoData));
    const max = _.max(Object.values(infoData));

    const scss = `
@font-face {
  font-family: "${font}";
  font-style: normal;
  font-weight: 900;
  src: url("../fonts/${font}.woff2") format("woff2");
  unicode-range: U+${min.toString(16).toUpperCase()}-${max.toString(16).toUpperCase()};
  font-display: block;
}
`;

    writeFileSync(`fonts/${font}.scss`, scss, "utf8");
  }
}

async function createFonts(fonts) {
  for (const font of fonts) {
    await createFont(font);
  }
}

const fonts = [
  "arcticons",
  "fa6-solid",
  "fa6-regular",
  "fa6-brands",
  "mdi",
  "fluent-emoji-high-contrast"
];

createFonts(fonts);
