<script lang="ts">
  import COLORS from "$lib/colors";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import Progress from "$lib/components/Progress.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { iconGlyph } from "$lib/icon";
  import { ajax, formatCurrency, helpUrl, type GoalSummary, formatPercentage } from "$lib/utils";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { onMount } from "svelte";

  let isEmpty = false;
  let goals: GoalSummary[] = [];

  onMount(async () => {
    ({ goals } = await ajax("/api/goals"));
    if (_.isEmpty(goals)) {
      isEmpty = true;
    }
  });

  function formatDate(date: string) {
    const d = dayjs(date, "YYYY-MM-DD", true);
    if (d.isValid()) {
      return d.fromNow();
    }
    return "";
  }

  function percentComplete(goal: GoalSummary) {
    if (goal.target === 0) {
      return 0;
    }

    return (goal.current / goal.target) * 100;
  }
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

      {#each goals as goal}
        {@const completed = percentComplete(goal)}
        <div class="column is-6 is-one-third-widescreen">
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
            <Progress small showPercent={false} progressPercent={completed} />
            <div class="flex justify-between has-text-grey">
              <div>{formatPercentage(completed / 100, 2)}</div>
              <div>{formatDate(goal.targetDate)}</div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>
