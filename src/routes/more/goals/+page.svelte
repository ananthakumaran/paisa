<script lang="ts">
  import COLORS from "$lib/colors";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import Progress from "$lib/components/Progress.svelte";
  import { iconGlyph } from "$lib/icon";
  import { ajax, formatCurrency, type GoalSummary } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let goals: GoalSummary[] = [];

  onMount(async () => {
    ({ goals } = await ajax("/api/goals"));
    console.log(goals);
  });

  function percentComplete(goal: GoalSummary) {
    return (goal.current / goal.target) * 100;
  }
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      {#each goals as goal}
        <div class="column is-one-third-widescreen is-half-desktop">
          <div class="box p-3">
            <div class="flex justify-between mb-4">
              <a
                class="secondary-link"
                href="/more/goals/{goal.type}/{encodeURIComponent(goal.name)}"
              >
                <h4 class="is-size-4">{goal.name}</h4>
              </a>
              {#if !_.isEmpty(goal.icon)}
                <span class="is-size-2 custom-icon">{iconGlyph(goal.icon)}</span>
              {/if}
            </div>
            <nav class="level grid-2">
              <LevelItem
                narrow
                title="Current"
                color={COLORS.gainText}
                value={formatCurrency(goal.current)}
              />

              <LevelItem
                narrow
                title="Target"
                color={COLORS.secondary}
                value={formatCurrency(goal.target)}
              />
            </nav>
            <Progress small progressPercent={percentComplete(goal)} />
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>
