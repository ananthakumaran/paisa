<script lang="ts">
  export let active = false;
  let close = function (_e: any) {
    active = !active;
  };
</script>

<div class="modal" class:is-active={active}>
  <div class="modal-background" on:click={(e) => close(e)} />
  <div class="modal-card">
    <header class="modal-card-head">
      <slot name="head" {close} />
    </header>
    <section class="modal-card-body">
      <slot name="body" />
    </section>
    <footer class="modal-card-foot">
      <slot name="foot" {close} />
    </footer>
  </div>
</div>
