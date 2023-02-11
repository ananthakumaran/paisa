<script lang="ts">
  export let title: string;
  export let value: string;
  export let color: string;
  export let subtitle: string;
  export let subvalue: string;
  export let subcolor: string;
</script>

<div class="level-item has-text-centered">
  <div>
    <p class="heading is-flex is-justify-content-space-between">
      <span>{title}</span>
      <span
        title={subtitle}
        class="tag ml-2 has-text-weight-semibold"
        style="padding: 5px; color: white; background-color: {subcolor};">{subvalue}</span
      >
    </p>
    <p class="title" style="padding: 5px; color: white; background-color: {color};">{value}</p>
  </div>
</div>
