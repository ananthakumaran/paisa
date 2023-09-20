<script lang="ts">
  import { onMount } from "svelte";
  import COLORS from "$lib/colors";
  import { ajax } from "$lib/utils";
  import { renderIssues } from "$lib/doctor";

  let issues = [];
  onMount(async () => {
    ({ issues } = await ajax("/api/diagnosis"));
    renderIssues(issues);
  });
</script>

<section class="section tab-doctor">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <b
            class="p-1 has-text-white"
            style="background-color: {issues.length > 0 ? COLORS.lossText : COLORS.gainText}"
            >{issues.length}</b
          > potential issue(s) found.
        </div>
      </div>
    </div>
    <div class="columns is-flex-wrap-wrap" id="d3-diagnosis" />
  </div>
</section>
