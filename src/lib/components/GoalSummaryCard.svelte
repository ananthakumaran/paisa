<script lang="ts">
  import { iconGlyph } from "$lib/icon";
  import { formatCurrency, formatPercentage, type GoalSummary } from "$lib/utils";
  import _ from "lodash";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import Progress from "$lib/components/Progress.svelte";
  import COLORS from "$lib/colors";
  import dayjs from "dayjs";

  export let goal: GoalSummary;
  export let small = false;

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

  $: completed = percentComplete(goal);
</script>

<div class="box p-3 goal-summary-card" class:mb-3={small}>
  <div class="flex justify-between mb-4">
    <a
      class="secondary-link has-text-grey"
      href="/more/goals/{goal.type}/{encodeURIComponent(goal.name)}"
    >
      <h4 class="is-size-4 has-text-grey">{goal.name}</h4>
    </a>
    {#if !_.isEmpty(goal.icon)}
      <span class="{small ? 'is-size-3' : 'is-size-2'} custom-icon">{iconGlyph(goal.icon)}</span>
    {/if}
  </div>
  <nav class="level grid-2">
    <LevelItem
      {small}
      narrow
      title="Current"
      color={COLORS.gainText}
      value={formatCurrency(goal.current)}
    />

    <LevelItem
      {small}
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
