import type { HandleClientError } from "@sveltejs/kit";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";
dayjs.extend(customParseFormat);
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
dayjs.extend(isSameOrBefore);
import isSameOrAfter from "dayjs/plugin/isSameOrAfter";
dayjs.extend(isSameOrAfter);
import relativeTime from "dayjs/plugin/relativeTime";
dayjs.extend(relativeTime, {
  thresholds: [
    { l: "s", r: 1 },
    { l: "m", r: 1 },
    { l: "mm", r: 59, d: "minute" },
    { l: "h", r: 1 },
    { l: "hh", r: 23, d: "hour" },
    { l: "d", r: 1 },
    { l: "dd", r: 29, d: "day" },
    { l: "M", r: 1 },
    { l: "MM", r: 11, d: "month" },
    { l: "y", r: 1 },
    { l: "yy", d: "year" },
  ],
});
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone"; // dependent on utc plugin
dayjs.extend(utc);
dayjs.extend(timezone);
import localeData from "dayjs/plugin/localeData";
dayjs.extend(localeData);
import updateLocale from "dayjs/plugin/updateLocale";
dayjs.extend(updateLocale);

import * as pdfjs from "pdfjs-dist";
// @ts-ignore: Vite worker ?url import
import pdfjsWorkerUrl from "pdfjs-dist/build/pdf.worker.js?url";

if (pdfjs.GlobalWorkerOptions) {
  pdfjs.GlobalWorkerOptions.workerSrc = pdfjsWorkerUrl;
}

import Handlebars from "handlebars";
import helpers from "$lib/template_helpers";
import * as toast from "bulma-toast";
import _ from "lodash";

import "@formatjs/intl-numberformat/polyfill";
import "@formatjs/intl-numberformat/locale-data/en";

Handlebars.registerHelper(
  _.mapValues(helpers, (helper, name) => {
    return function (this: any, ...args: any[]) {
      try {
        return helper.apply(this, args);
      } catch (e) {
        console.log("Error in helper", name, args, e);
      }
    };
  }),
);

toast.setDefaults({
  position: "bottom-right",
  dismissible: true,
  pauseOnHover: true,
  extraClasses: "is-light invertable",
} as any);

globalThis.USER_CONFIG = {} as any;

export const handleError: HandleClientError = (
  { error, status, message },
) => {
  let stack: string | undefined = undefined;
  if (error instanceof Error) {
    stack = error.stack;
  }
  return { message, stack, status, detail: (error as any)?.toString?.() ?? String(error) } as any;
};

function formatError(error: any) {
  if (error?.stack) {
    return error.stack;
  }

  if (error?.message) {
    return error.message;
  } else {
    return String(error);
  }
}

const footer = `
<p class="mt-3">
  Please report this issue at <a href="https://github.com/ananthakumaran/paisa/issues"
    >https://github.com/ananthakumaran/paisa/issues</a
  >. Closing and reopening the app may help.
</p>
`;

function displayError(error: any) {
  const message = formatError(error);
  toast.toast({
    message:
      `<div class="message invertable is-danger"><div class="message-header">Something Went Wrong</div><div class="message-body">${message}${footer}</div></div>`,
    type: "is-danger",
    dismissible: true,
    pauseOnHover: true,
    duration: 10000,
    position: "center",
    animate: { in: "fadeIn", out: "fadeOut" },
  });
}

globalThis.addEventListener("unhandledrejection", (event: any) => {
  displayError(event.reason);
});
globalThis.addEventListener("error", (event: any) => {
  displayError(event.error);
});
