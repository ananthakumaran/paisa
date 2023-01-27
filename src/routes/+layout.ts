export const prerender = true;
export const ssr = false;

import "@fortawesome/fontawesome-free/css/all.css";
import "../app.scss";
import "tippy.js/dist/tippy.css";
import "tippy.js/themes/light.css";

import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
dayjs.extend(isSameOrBefore);
