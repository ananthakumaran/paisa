import $ from "jquery";
import _ from "lodash";
import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
dayjs.extend(isSameOrBefore);

import overview from "./overview";
import investment from "./investment";
import ledger from "./ledger";

const tabs = {
  overview: _.once(overview),
  investment: _.once(investment),
  ledger: _.once(ledger)
};

function toggleTab(id: string) {
  const ids = _.keys(tabs);
  _.each(ids, (tab) => {
    $(`section.tab-${tab}`).hide();
  });
  $(`section.tab-${id}`).show();
  tabs[id]();
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
