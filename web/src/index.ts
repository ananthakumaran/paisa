// css

import "clusterize.js/clusterize.css";
import "bulma/css/bulma.css";
import "../static/styles/custom.css";

import { followCursor, Instance, delegate } from "tippy.js";
import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
import $ from "jquery";
import _ from "lodash";
dayjs.extend(isSameOrBefore);
import "tippy.js/dist/tippy.css";
import "tippy.js/themes/light.css";

import allocation from "./allocation";
import investment from "./investment";
import holding from "./holding";
import journal from "./journal";
import overview from "./overview";
import gain from "./gain";
import income from "./income";
import expense from "./expense";
import price from "./price";
import harvest from "./harvest";
import doctor from "./doctor";

const tabs = {
  overview: _.once(overview),
  investment: _.once(investment),
  allocation: _.once(allocation),
  holding: _.once(holding),
  journal: _.once(journal),
  gain: _.once(gain),
  income: _.once(income),
  expense: _.once(expense),
  price: _.once(price),
  harvest: _.once(harvest),
  doctor: _.once(doctor)
};

let tippyInstances: Instance[] = [];

function toggleTab(id: string) {
  const ids = _.keys(tabs);
  _.each(ids, (tab) => {
    $(`section.tab-${tab}`).hide();
  });
  $(`section.tab-${id}`).show();
  $(window).scrollTop(0);
  tabs[id]().then(function () {
    tippyInstances.forEach((t) => t.destroy());
    tippyInstances = delegate(`section.tab-${id}`, {
      target: "[data-tippy-content]",
      theme: "light",
      onShow: (instance) => {
        instance.setContent(
          instance.reference.getAttribute("data-tippy-content")
        );
      },
      maxWidth: "none",
      delay: 0,
      allowHTML: true,
      followCursor: true,
      plugins: [followCursor]
    });
  });
}

$("a.navbar-item").on("click", function () {
  const id = $(this).attr("id");
  window.location.hash = id;
  toggleTab(id);
  $(".navbar-item, .navbar-link").removeClass("is-active");
  $(this).addClass("is-active");
  $(this)
    .closest(".has-dropdown.navbar-item")
    .find(".navbar-link")
    .addClass("is-active");
  return false;
});

if (!_.isEmpty(window.location.hash)) {
  $(window.location.hash).trigger("click");
} else {
  $("#overview").trigger("click");
}

$(".navbar-burger").on("click", function () {
  $(".navbar-burger").toggleClass("is-active");
  $(".navbar-menu").toggleClass("is-active");
});
