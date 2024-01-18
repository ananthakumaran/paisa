<script lang="ts">
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import GoalSummaryCard from "$lib/components/GoalSummaryCard.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { ajax, helpUrl, type GoalSummary } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import * as toast from "bulma-toast";

  let isEmpty = false;
  let config: UserConfig;
  let goals: GoalSummary[] = [];

  function handleConsider(event: CustomEvent<DndEvent<GoalSummary>>) {
    goals = event.detail.items;
  }

  async function handleFinalize(event: CustomEvent<DndEvent<GoalSummary>>) {
    goals = event.detail.items;
    for (let i = 0; i < goals.length; i++) {
      const g = goals[i];
      g.priority = goals.length - i;
      const goalConfig = _.find(config.goals[g.type] || [], { name: g.name });
      if (goalConfig) {
        goalConfig.priority = g.priority;
      }
    }
    await save(config);
  }

  async function save(newConfig: UserConfig) {
    const { success, error } = await ajax("/api/config", {
      method: "POST",
      body: JSON.stringify(newConfig),
      background: true
    });

    if (success) {
      globalThis.USER_CONFIG = _.cloneDeep(newConfig);
      toast.toast({
        message: `Updated goal config`,
        type: "is-success"
      });
    } else {
      toast.toast({
        message: `Failed to save config: ${error}`,
        type: "is-danger",
        duration: 10000
      });
    }
  }

  onMount(async () => {
    ({ config } = await ajax("/api/config"));
    ({ goals } = await ajax("/api/goals"));
    goals = _.sortBy(goals, (g) => -g.priority);
    if (_.isEmpty(goals)) {
      isEmpty = true;
    }
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <strong>Oops!</strong> You haven't configured any goals yet. Checkout the
          <a href={helpUrl("goals")}>docs</a> page to get started.
        </ZeroState>
      </div>
    </div>
    <div
      class="columns flex-wrap"
      use:dndzone={{ items: goals, dropTargetStyle: {}, flipDurationMs: 300 }}
      on:consider={handleConsider}
      on:finalize={handleFinalize}
    >
      {#each goals as goal (goal.id)}
        <div animate:flip={{ duration: 300 }} class="column is-6 is-one-third-widescreen">
          <GoalSummaryCard {goal} />
        </div>
      {/each}
    </div>
  </div>
</section>
