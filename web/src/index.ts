import tippy, { followCursor, Instance } from "tippy.js";
import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
import $ from "jquery";
import _ from "lodash";
dayjs.extend(isSameOrBefore);
import "tippy.js/dist/tippy.css";
import "tippy.js/themes/light.css";

import allocation from "./allocation";
import investment from "./investment";
import ledger from "./ledger";
import overview from "./overview";
import gain from "./gain";
import income from "./income";
import expense from "./expense";

const tabs = {
  overview: _.once(overview),
  investment: _.once(investment),
  allocation: _.once(allocation),
  ledger: _.once(ledger),
  gain: _.once(gain),
  income: _.once(income),
  expense: _.once(expense)
};

let tippyInstances: Instance[] = [];

function toggleTab(id: string) {
  const ids = _.keys(tabs);
  _.each(ids, (tab) => {
    $(`section.tab-${tab}`).hide();
  });
  $(`section.tab-${id}`).show();
  tabs[id]().then(function () {
    tippyInstances.forEach((t) => t.destroy());
    tippyInstances = tippy(`section.tab-${id} [data-tippy-content]`, {
      theme: "light",
      delay: 0,
      allowHTML: true,
      followCursor: true,
      plugins: [followCursor]
    });
  });
}

$("a.navbar-item").on("click", function () {
  const id = $(this).attr("id");
  toggleTab(id);
  window.location.hash = id;
  $(".navbar-item").removeClass("is-active");
  $(this).addClass("is-active");
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
