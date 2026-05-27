<script lang="ts">
  import { onMount } from "svelte";
  import { ajax } from "$lib/utils";
  import type { Issue, Severity } from "$lib/utils";
  import { groupBySeverity, bulmaClassFor, severityLabel, SEVERITY_ORDER } from "$lib/doctor";

  let issues: Issue[] = [];
  // Infos are collapsed by default — they are informational, not blocking.
  let collapsed: Record<Severity, boolean> = {
    error: false,
    warning: false,
    info: true
  };

  onMount(async () => {
    ({ issues } = await ajax("/api/diagnosis"));
  });

  $: groups = groupBySeverity(issues);
  $: counts = SEVERITY_ORDER.reduce(
    (acc, sev) => {
      acc[sev] = groups.find((g) => g.severity === sev)?.issues.length ?? 0;
      return acc;
    },
    { error: 0, warning: 0, info: 0 } as Record<Severity, number>
  );

  function toggle(severity: Severity) {
    collapsed[severity] = !collapsed[severity];
  }
</script>

<section class="section tab-doctor">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        {#if issues.length === 0}
          <div>
            <b class="tag is-success">0</b> potential issue(s) found.
          </div>
        {:else}
          <div class="severity-counter">
            <span class="tag is-danger">{counts.error}</span>
            <span class="counter-label">{severityLabel("error", counts.error)}</span>
            <span class="counter-sep">/</span>
            <span class="tag is-warning">{counts.warning}</span>
            <span class="counter-label">{severityLabel("warning", counts.warning)}</span>
            <span class="counter-sep">/</span>
            <span class="tag is-info">{counts.info}</span>
            <span class="counter-label">{severityLabel("info", counts.info)}</span>
          </div>
        {/if}
      </div>
    </div>

    {#each groups as group (group.severity)}
      <div class="columns">
        <div class="column is-12">
          <div
            class="severity-group-header"
            class:is-collapsed={collapsed[group.severity]}
            role="button"
            tabindex="0"
            on:click={() => toggle(group.severity)}
            on:keydown={(e) => (e.key === "Enter" || e.key === " ") && toggle(group.severity)}
          >
            <span class="icon">
              <i
                class="fas"
                class:fa-chevron-down={!collapsed[group.severity]}
                class:fa-chevron-right={collapsed[group.severity]}
              />
            </span>
            <span class="tag {bulmaClassFor(group.severity)}">{group.issues.length}</span>
            <span class="severity-title">{severityLabel(group.severity, group.issues.length)}</span>
          </div>
        </div>
      </div>

      {#if !collapsed[group.severity]}
        <div class="columns is-flex-wrap-wrap">
          {#each group.issues as issue}
            <div class="column is-6">
              <div
                class="message invertable {bulmaClassFor(group.severity)}"
                class:severity-info-muted={group.severity === "info"}
              >
                <div class="message-header">
                  <p>{issue.summary}</p>
                </div>
                <div class="message-body">
                  {@html issue.description}
                  {#if issue.details}
                    <br /><br />
                    {@html issue.details}
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    {/each}
  </div>
</section>

<style>
  .severity-counter {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex-wrap: wrap;
    justify-content: center;
  }
  .counter-label {
    font-weight: 500;
  }
  .counter-sep {
    opacity: 0.4;
    margin: 0 0.25rem;
  }
  .severity-group-header {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    padding: 0.4rem 0.6rem;
    border-radius: 4px;
    user-select: none;
  }
  .severity-group-header:hover {
    background-color: rgba(0, 0, 0, 0.04);
  }
  .severity-title {
    text-transform: capitalize;
    font-weight: 600;
  }
  .severity-info-muted .message-header {
    opacity: 0.85;
  }
  .severity-info-muted {
    opacity: 0.9;
  }
</style>
