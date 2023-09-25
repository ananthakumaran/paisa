<script lang="ts">
  import { onMount } from "svelte";
  import { ajax, type Log } from "$lib/utils";
  import VirtualList from "svelte-tiny-virtual-list";
  import _ from "lodash";

  let logs: Log[] = [];
  const ITEM_SIZE = 20;
  onMount(async () => {
    ({ logs } = await ajax("/api/logs"));
  });

  function levelClass(level: string) {
    switch (level) {
      case "info":
        return "is-info";
      case "warning":
        return "is-warning";
      case "error":
      case "fatal":
        return "is-danger";
      default:
        return "";
    }
  }
  function formatFields(log: Log) {
    return _.map(_.omit(log, ["time", "level", "msg"]), (value, key) => `${key}=${value}`).join(
      ", "
    );
  }
</script>

<section class="section tab-price">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="columns">
          <div class="column is-12">
            <div class="box px-2">
              <VirtualList
                width="100%"
                height={window.innerHeight - 130}
                itemCount={logs.length}
                itemSize={ITEM_SIZE}
              >
                <div slot="item" let:index let:style {style}>
                  {@const log = logs[index]}
                  {@const fields = _.omit(log, ["time", "level", "msg"])}
                  <div class="is-flex log is-align-items-baseline">
                    <div class="time is-size-7">{log.time.format("YYYY-MM-DD HH:mm:ss")}</div>
                    <div
                      class="is-size-7 tag is-small is-light invertable py-0 log-level {levelClass(
                        log.level
                      )}"
                    >
                      {log.level}
                    </div>
                    <div class="msg truncate" title={log.msg}>{log.msg}</div>
                    <div class="fields is-size-7 truncate" title={formatFields(log)}>
                      {#each Object.entries(fields) as [key, value]}
                        <span class="px-1 field"><span>{key}</span>=<span>{value}</span></span>
                      {/each}
                    </div>
                  </div>
                </div>
              </VirtualList>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .log {
    gap: 5px;

    .time {
      width: 125px;
      flex-shrink: 0;
    }
    div {
      padding: 0 5px;
    }

    .log-level {
      text-transform: uppercase;
      width: 60px;
      height: 1.4em !important;
    }

    .msg {
      width: 500px;
    }

    .fields {
      font-family: $family-monospace;
      flex-basis: 20px;
      flex-grow: 1;
      flex-shrink: 1;
    }
  }
</style>
