<script lang="ts">
  import { onMount } from "svelte";
  import COLORS from "$lib/colors";
  import { ajax, setHtml } from "$lib/utils";
  import { renderIssues } from "$lib/doctor";

  onMount(async () => {
    const { issues: issues } = await ajax("/api/diagnosis");
    setHtml(
      "diagnosis-count",
      issues.length.toString(),
      issues.length > 0 ? COLORS.lossText : COLORS.gainText
    );
    renderIssues(issues);
  });
</script>

<section class="section tab-doctor">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          We have found <b class="d3-diagnosis-count" /> potential issue(s).
        </div>
      </div>
    </div>
    <div class="columns is-flex-wrap-wrap" id="d3-diagnosis" />
  </div>
</section>
