<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let selectedMonths: string[] = [];

  const dispatch = createEventDispatcher();

  const months = [
    { value: "01", label: "January" },
    { value: "02", label: "February" },
    { value: "03", label: "March" },
    { value: "04", label: "April" },
    { value: "05", label: "May" },
    { value: "06", label: "June" },
    { value: "07", label: "July" },
    { value: "08", label: "August" },
    { value: "09", label: "September" },
    { value: "10", label: "October" },
    { value: "11", label: "November" },
    { value: "12", label: "December" }
  ];

  let isOpen = false;
  let dropdownElement: HTMLDivElement;

  function toggleDropdown(event: MouseEvent) {
    event.stopPropagation();
    isOpen = !isOpen;
  }

  function toggleMonth(monthValue: string) {
    if (selectedMonths.includes(monthValue)) {
      selectedMonths = selectedMonths.filter((m) => m !== monthValue);
    } else {
      selectedMonths = [...selectedMonths, monthValue];
    }
    dispatch("change", selectedMonths);
  }

  function selectAllMonths(event: MouseEvent) {
    event.stopPropagation();
    selectedMonths = months.map((m) => m.value);
    dispatch("change", selectedMonths);
  }

  function clearAllMonths(event: MouseEvent) {
    event.stopPropagation();
    // Select just January to keep it predictable and avoid auto-selection
    selectedMonths = ["01"];
    dispatch("change", selectedMonths);
  }

  function handleClickOutside(event: MouseEvent) {
    if (dropdownElement && !dropdownElement.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  $: selectedMonthsText = (() => {
    if (selectedMonths.length === 0) return "Select months";
    if (selectedMonths.length === 1) {
      const month = months.find((m) => m.value === selectedMonths[0]);
      return month ? month.label : selectedMonths[0];
    }
    if (selectedMonths.length === 12) return "All months";
    return `${selectedMonths.length} months selected`;
  })();
</script>

<svelte:window on:click={handleClickOutside} />

<div class="dropdown is-left is-small" class:is-active={isOpen} bind:this={dropdownElement}>
  <div class="dropdown-trigger">
    <button
      class="button is-small"
      on:click={toggleDropdown}
      type="button"
      aria-haspopup="true"
      aria-controls="dropdown-menu"
    >
      <span>{selectedMonthsText}</span>
      <span class="icon is-small">
        <i class="fas fa-chevron-down" class:rotated={isOpen}></i>
      </span>
    </button>
  </div>

  <div class="dropdown-menu" id="dropdown-menu" role="menu">
    <div class="dropdown-content">
      <!-- Select All / Clear All buttons -->
      <div class="dropdown-actions">
        <button class="action-button select-all" on:click={selectAllMonths} type="button">
          Select All
        </button>
        <button class="action-button clear-all" on:click={clearAllMonths} type="button">
          Clear All
        </button>
      </div>

      <div class="dropdown-divider"></div>

      <!-- Month checkboxes -->
      {#each months as month}
        <label class="dropdown-item checkbox-item">
          <input
            type="checkbox"
            checked={selectedMonths.includes(month.value)}
            on:change={() => toggleMonth(month.value)}
          />
          <span class="checkbox-label">{month.label}</span>
          <span class="month-code">({month.value})</span>
        </label>
      {/each}
    </div>
  </div>
</div>

<style>
  /* Icon rotation animation */
  .dropdown-trigger .icon i.rotated {
    transform: rotate(180deg);
  }

  /* Custom dropdown content styling */
  .dropdown-content {
    max-height: 320px;
    overflow-y: auto;
  }

  .dropdown-actions {
    display: flex;
    padding: 0.75rem;
    gap: 0.5rem;
    background: #f8f9fa;
    border-bottom: 0px solid #e9ecef;
  }

  .action-button {
    flex: 1;
    padding: 0.375rem 0.75rem;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    background: white;
    color: #495057;
    font-size: 0.75rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-button:hover {
    background: #e9ecef;
    border-color: #adb5bd;
  }

  .action-button.select-all:hover {
    background: #d4edda;
    border-color: #c3e6cb;
    color: #155724;
  }

  .action-button.clear-all:hover {
    background: #f8d7da;
    border-color: #f5c6cb;
    color: #721c24;
  }

  .dropdown-divider {
    height: 1px;
    background: #e9ecef;
    margin: 0;
  }

  .checkbox-item {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    cursor: pointer;
    transition: background-color 0.2s ease;
    margin: 0;
    border-bottom: 0px solid #f8f9fa;
  }

  .checkbox-item:last-child {
    border-bottom: none;
  }

  .checkbox-item:hover {
    background-color: #f8f9fa;
  }

  .checkbox-item input[type="checkbox"] {
    margin-right: 0.75rem;
    cursor: pointer;
    width: 16px;
    height: 16px;
    accent-color: #3273dc;
  }

  .checkbox-label {
    cursor: pointer;
    user-select: none;
    flex: 1;
    font-weight: 500;
  }

  .month-code {
    color: #6c757d;
    font-size: 0.75rem;
    font-weight: normal;
  }

  /* Dark mode support */
  :global(html[data-theme="dark"]) .dropdown-trigger {
    background: #363636 !important;
    border-color: #4a4a4a !important;
    color: #f5f5f5 !important;
  }

  :global(html[data-theme="dark"]) .dropdown-trigger:hover {
    border-color: #5a5a5a !important;
  }

  :global(html[data-theme="dark"]) .dropdown-menu {
    background: #363636 !important;
    border-color: #4a4a4a !important;
  }

  :global(html[data-theme="dark"]) .dropdown-actions {
    background: #2a2a2a !important;
    border-bottom: 0px solid #4a4a4a !important;
  }

  :global(html[data-theme="dark"]) .action-button {
    background: #363636 !important;
    border-color: #4a4a4a !important;
    color: #f5f5f5 !important;
  }

  :global(html[data-theme="dark"]) .action-button:hover {
    background: #4a4a4a !important;
    border-color: #5a5a5a !important;
  }

  :global(html[data-theme="dark"]) .action-button.select-all:hover {
    background: #1e4d2b !important;
    border-color: #28a745 !important;
    color: #d4edda !important;
  }

  :global(html[data-theme="dark"]) .action-button.clear-all:hover {
    background: #4d1e1e !important;
    border-color: #dc3545 !important;
    color: #f8d7da !important;
  }

  :global(html[data-theme="dark"]) .dropdown-divider {
    background: #4a4a4a !important;
  }

  :global(html[data-theme="dark"]) .checkbox-item {
    border-bottom: 0px solid #4a4a4a !important;
    color: #f5f5f5 !important;
  }

  :global(html[data-theme="dark"]) .checkbox-item:hover {
    background-color: #2a2a2a !important;
  }

  :global(html[data-theme="dark"]) .month-code {
    color: #adb5bd !important;
  }
</style>
