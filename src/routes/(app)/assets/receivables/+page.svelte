<script lang="ts">
  import Table from "$lib/components/Table.svelte";
  import { _ as t } from "$lib/i18n";
  import { sortByOutstandingDesc, type Receivable } from "$lib/receivables";
  import { ajax, formatCurrency, now } from "$lib/utils";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { onMount } from "svelte";
  import type { CellComponent, ColumnDefinition } from "tabulator-tables";

  let receivables: Receivable[] = [];
  let total = 0;
  let isEmpty = false;

  onMount(async () => {
    const resp = await ajax("/api/receivables");
    receivables = sortByOutstandingDesc(resp.receivables ?? []);
    total = resp.total_outstanding ?? 0;
    isEmpty = receivables.length === 0;
  });

  function formatDate(cell: CellComponent): string {
    const v = cell.getValue() as dayjs.Dayjs | null | undefined;
    if (!v) {
      return "";
    }
    return v.format("YYYY/MM/DD");
  }

  function formatDueDate(cell: CellComponent): string {
    const v = cell.getValue() as dayjs.Dayjs | null | undefined;
    if (!v) {
      return "";
    }
    const isOverdue = v.isBefore(now(), "day");
    const cls = isOverdue ? "has-text-danger has-text-weight-semibold" : "";
    return `<span class="${cls}">${v.format("YYYY/MM/DD")}</span>`;
  }

  function formatInterestRate(cell: CellComponent): string {
    const v = Number(cell.getValue() ?? 0);
    if (!v) {
      return "—";
    }
    // Already expressed as APR percent in the config (e.g. 4.9 -> 4.9%).
    return `${v}%`;
  }

  function formatOutstanding(cell: CellComponent): string {
    const v = Number(cell.getValue() ?? 0);
    return formatCurrency(v);
  }

  function formatBorrower(cell: CellComponent): string {
    const v = (cell.getValue() ?? "") as string;
    return _.escape(v);
  }

  function formatAccount(cell: CellComponent): string {
    const v = (cell.getValue() ?? "") as string;
    return `<a href="/assets/gain/${encodeURIComponent(v)}">${_.escape(v)}</a>`;
  }

  function formatNote(cell: CellComponent): string {
    return _.escape((cell.getValue() ?? "") as string);
  }

  $: columns = [
    {
      title: $t("receivables.column.borrower"),
      field: "borrower",
      formatter: formatBorrower,
      frozen: true
    },
    {
      title: $t("receivables.column.account"),
      field: "account",
      formatter: formatAccount
    },
    {
      title: $t("receivables.column.outstanding"),
      field: "outstanding",
      hozAlign: "right",
      formatter: formatOutstanding,
      sorter: "number"
    },
    {
      title: $t("receivables.column.lend_date"),
      field: "lend_date",
      hozAlign: "right",
      formatter: formatDate
    },
    {
      title: $t("receivables.column.due_date"),
      field: "due_date",
      hozAlign: "right",
      formatter: formatDueDate
    },
    {
      title: $t("receivables.column.interest_rate"),
      field: "interest_rate",
      hozAlign: "right",
      formatter: formatInterestRate
    },
    {
      title: $t("receivables.column.note"),
      field: "note",
      formatter: formatNote
    }
  ] satisfies ColumnDefinition[];
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-6 has-text-centered">
        <article class="message">
          <div class="message-body">
            {$t("receivables.empty")}
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section" class:is-hidden={isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <div class="level mb-0">
            <div class="level-left">
              <div class="level-item">
                <div>
                  <p class="heading">{$t("receivables.summary.total_outstanding")}</p>
                  <p class="title is-4">{formatCurrency(total)}</p>
                </div>
              </div>
            </div>
            <div class="level-right">
              <div class="level-item">
                <div class="has-text-right">
                  <p class="heading">{$t("receivables.summary.count")}</p>
                  <p class="title is-4">{receivables.length}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12">
        <Table data={receivables} {columns} />
      </div>
    </div>
  </div>
</section>
