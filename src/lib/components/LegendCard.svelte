<script lang="ts">
  import * as d3 from "d3";
  import type { Action } from "svelte/action";
  import type { Legend } from "$lib/utils";

  export let clazz = "";
  export let legends: Legend[];

  const textureScale = 14;
  const texture: Action<SVGSVGElement, { texture: any }> = (element, props) => {
    const svg = d3.select(element);
    svg.call(props.texture);
    svg
      .append("rect")
      .attr("x", 0)
      .attr("y", 0)
      .attr("height", textureScale)
      .attr("width", textureScale)
      .attr("fill", props.texture.url());

    return {};
  };

  let selectedLegend: Legend;

  function onClick(legend: Legend) {
    if (!legend.onClick) {
      return;
    }

    legend.onClick(legend);
    if (selectedLegend == legend) {
      // toggle
      legend.selected = false;
      selectedLegend = null;
    } else {
      selectedLegend && (selectedLegend.selected = false);
      legend.selected = true;
      selectedLegend = legend;
    }
  }
</script>

<div class="flex justify-start gap-0 {clazz}">
  {#each legends as legend}
    <div
      class="flex flex-col p-1.5 gap-2 legend-box {legend.onClick && 'cursor-pointer'}"
      on:click={(_e) => onClick(legend)}
      class:selected={selectedLegend == legend}
    >
      {#if legend.texture}
        <svg
          use:texture={{ texture: legend.texture }}
          class="self-center"
          height="1rem"
          width="1rem"
          viewBox="0 0 {textureScale} {textureScale}"
        ></svg>
      {:else if legend.shape == "square"}
        <div
          class="self-center"
          style="background-color: {legend.color}; height: 1rem; width: 1rem;"
        ></div>
      {:else if legend.shape == "line"}
        <div
          class="self-center"
          style="border-top: 3px solid {legend.color}; height: 0.1rem; width: 2rem;"
        ></div>
      {/if}
      <div class="legend-label whitespace-pre is-size-6-5 has-text-grey custom-icon">
        {legend.label}
      </div>
    </div>
  {/each}
</div>
