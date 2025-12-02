<script lang="ts">
  import { onMount } from "svelte";
  import { ajax } from "$lib/utils";
  import { accountRules, type AccountRule } from "../../../../store";
  import _ from "lodash";
  import * as toast from "bulma-toast";
  import Modal from "$lib/components/Modal.svelte";

  let rules: AccountRule[] = [];
  let editingRule: AccountRule | null = null;
  let modalOpen = false;
  let isNewRule = false;

  const defaultRule: AccountRule = {
    name: "",
    pattern: "",
    account: "",
    description: "",
    enabled: true
  };

  onMount(async () => {
    await loadRules();
  });

  async function loadRules() {
    try {
      const response = await ajax("/api/account-rules");
      rules = response.rules || [];
      accountRules.set(rules);
    } catch (error) {
      console.error("Failed to load account rules:", error);
      toast.toast({
        message: "Failed to load account rules",
        type: "is-danger"
      });
    }
  }

  function openCreateModal() {
    editingRule = _.cloneDeep(defaultRule);
    isNewRule = true;
    modalOpen = true;
  }

  function openEditModal(rule: AccountRule) {
    editingRule = _.cloneDeep(rule);
    isNewRule = false;
    modalOpen = true;
  }

  async function saveRule() {
    if (!editingRule || !editingRule.name.trim() || !editingRule.pattern.trim() || !editingRule.account.trim()) {
      toast.toast({
        message: "Please fill in all required fields",
        type: "is-danger"
      });
      return;
    }

    // Test the regex pattern
    try {
      new RegExp(editingRule.pattern);
    } catch (e) {
      toast.toast({
        message: "Invalid regex pattern",
        type: "is-danger"
      });
      return;
    }

    try {
      const { saved, message } = await ajax("/api/account-rules/upsert", {
        method: "POST",
        body: JSON.stringify(editingRule),
        background: true
      });

      if (saved) {
        toast.toast({
          message: `${isNewRule ? "Created" : "Updated"} rule "${editingRule.name}"`,
          type: "is-success"
        });
        await loadRules();
        modalOpen = false;
        editingRule = null;
      } else {
        toast.toast({
          message: `Failed to save rule: ${message}`,
          type: "is-danger",
          duration: 10000
        });
      }
    } catch (error) {
      console.error("Failed to save rule:", error);
      toast.toast({
        message: "Failed to save rule",
        type: "is-danger"
      });
    }
  }

  async function deleteRule(rule: AccountRule) {
    if (!confirm(`Are you sure you want to delete rule "${rule.name}"?`)) {
      return;
    }

    try {
      const { success, message } = await ajax("/api/account-rules/delete", {
        method: "POST",
        body: JSON.stringify(rule),
        background: true
      });

      if (success) {
        toast.toast({
          message: `Deleted rule "${rule.name}"`,
          type: "is-success"
        });
        await loadRules();
      } else {
        toast.toast({
          message: `Failed to delete rule: ${message}`,
          type: "is-danger",
          duration: 10000
        });
      }
    } catch (error) {
      console.error("Failed to delete rule:", error);
      toast.toast({
        message: "Failed to delete rule",
        type: "is-danger"
      });
    }
  }

  async function toggleRule(rule: AccountRule) {
    const updatedRule = { ...rule, enabled: !rule.enabled };
    
    try {
      const { saved, message } = await ajax("/api/account-rules/upsert", {
        method: "POST",
        body: JSON.stringify(updatedRule),
        background: true
      });

      if (saved) {
        toast.toast({
          message: `${updatedRule.enabled ? "Enabled" : "Disabled"} rule "${rule.name}"`,
          type: "is-success"
        });
        await loadRules();
      } else {
        toast.toast({
          message: `Failed to update rule: ${message}`,
          type: "is-danger"
        });
      }
    } catch (error) {
      console.error("Failed to toggle rule:", error);
      toast.toast({
        message: "Failed to update rule",
        type: "is-danger"
      });
    }
  }

  function testPattern() {
    if (!editingRule || !editingRule.pattern.trim()) {
      return;
    }

    const testInput = prompt("Enter test text to match against the pattern:");
    if (testInput === null) return;

    try {
      const regex = new RegExp(editingRule.pattern, "i");
      const matches = regex.test(testInput);
      
      toast.toast({
        message: matches ? "✓ Pattern matches!" : "✗ Pattern does not match",
        type: matches ? "is-success" : "is-warning"
      });
    } catch (e) {
      toast.toast({
        message: "Invalid regex pattern",
        type: "is-danger"
      });
    }
  }
</script>

<svelte:head>
  <title>Account Rules</title>
</svelte:head>

<Modal bind:active={modalOpen}>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">{isNewRule ? "Create" : "Edit"} Account Rule</p>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  
  <div slot="body" class="content">
    {#if editingRule}
      <div class="field">
        <label class="label" for="rule-name">Rule Name *</label>
        <div class="control">
          <input 
            id="rule-name"
            class="input" 
            type="text" 
            bind:value={editingRule.name} 
            placeholder="e.g., Amazon Purchases"
          />
        </div>
      </div>

      <div class="field">
        <label class="label" for="rule-pattern">Regex Pattern *</label>
        <div class="control">
          <input 
            id="rule-pattern"
            class="input" 
            type="text" 
            bind:value={editingRule.pattern} 
            placeholder="e.g., AMAZON.*|AMZ.*"
          />
        </div>
        <p class="help">
          Regular expression to match against transaction data. Case-insensitive by default.
          <button type="button" class="button is-small is-link is-light ml-2" on:click={testPattern}>
            Test Pattern
          </button>
        </p>
      </div>

      <div class="field">
        <label class="label" for="rule-account">Account *</label>
        <div class="control">
          <input 
            id="rule-account"
            class="input" 
            type="text" 
            bind:value={editingRule.account} 
            placeholder="e.g., Expenses:Shopping:Online"
          />
        </div>
      </div>

      <div class="field">
        <label class="label" for="rule-description">Description</label>
        <div class="control">
          <textarea 
            id="rule-description"
            class="textarea" 
            bind:value={editingRule.description}
            placeholder="Optional description of what this rule does"
            rows="3"
          />
        </div>
      </div>

      <div class="field">
        <div class="control">
          <label class="checkbox">
            <input type="checkbox" bind:checked={editingRule.enabled} />
            Enabled
          </label>
        </div>
      </div>
    {/if}
  </div>

  <svelte:fragment slot="foot" let:close>
    <button class="button is-success" on:click={saveRule}>
      {isNewRule ? "Create" : "Update"} Rule
    </button>
    <button class="button" on:click={(e) => close(e)}>Cancel</button>
  </svelte:fragment>
</Modal>

<section class="section">
  <div class="container is-fluid">
    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <h1 class="title">Account Rules</h1>
        </div>
      </div>
      <div class="level-right">
        <div class="level-item">
          <button class="button is-primary" on:click={openCreateModal}>
            <span class="icon">
              <i class="fas fa-plus" />
            </span>
            <span>New Rule</span>
          </button>
        </div>
      </div>
    </div>

    <div class="box">
      <div class="content">
        <p>
          Account rules use regular expressions to automatically predict accounts based on transaction data.
          Rules are processed in order, and the first matching rule will be used.
        </p>
        <p class="has-text-weight-semibold">
          To use these rules in import templates, replace <code>predictAccount</code> with <code>predictAccountWithRules</code>.
        </p>
      </div>
    </div>

    {#if rules.length === 0}
      <div class="box has-text-centered">
        <p class="title is-4 has-text-grey-light">No Account Rules</p>
        <p class="subtitle has-text-grey">
          Create your first rule to start automatically predicting accounts during import.
        </p>
        <button class="button is-primary is-large" on:click={openCreateModal}>
          <span class="icon">
            <i class="fas fa-plus" />
          </span>
          <span>Create First Rule</span>
        </button>
      </div>
    {:else}
      <div class="table-container">
        <table class="table is-fullwidth is-striped is-hoverable">
          <thead>
            <tr>
              <th>Name</th>
              <th>Pattern</th>
              <th>Account</th>
              <th>Description</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each rules as rule (rule.name)}
              <tr class:has-background-grey-lighter={!rule.enabled}>
                <td>
                  <strong class:has-text-grey-light={!rule.enabled}>{rule.name}</strong>
                </td>
                <td>
                  <code class="is-size-7 {rule.enabled ? '' : 'has-text-grey-light'}">{rule.pattern}</code>
                </td>
                <td class:has-text-grey-light={!rule.enabled}>
                  {rule.account}
                </td>
                <td class="is-size-7 {rule.enabled ? 'has-text-grey' : 'has-text-grey-light'}">
                  {rule.description || '-'}
                </td>
                <td>
                  <span class="tag {rule.enabled ? 'is-success' : 'is-light'}">
                    {rule.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </td>
                <td>
                  <div class="field is-grouped">
                    <p class="control">
                      <button 
                        class="button is-small"
                        on:click={() => toggleRule(rule)}
                        title={rule.enabled ? 'Disable rule' : 'Enable rule'}
                      >
                        <span class="icon">
                          <i class="fas {rule.enabled ? 'fa-pause' : 'fa-play'}" />
                        </span>
                      </button>
                    </p>
                    <p class="control">
                      <button 
                        class="button is-small is-info" 
                        on:click={() => openEditModal(rule)}
                        title="Edit rule"
                      >
                        <span class="icon">
                          <i class="fas fa-edit" />
                        </span>
                      </button>
                    </p>
                    <p class="control">
                      <button 
                        class="button is-small is-danger" 
                        on:click={() => deleteRule(rule)}
                        title="Delete rule"
                      >
                        <span class="icon">
                          <i class="fas fa-trash" />
                        </span>
                      </button>
                    </p>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</section>

<style>
  .help button {
    vertical-align: baseline;
  }
</style>