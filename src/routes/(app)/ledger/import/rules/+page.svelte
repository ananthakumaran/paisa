<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { toast } from "$lib/utils";
  import { obscure } from "../../../../../persisted_store";

  interface Rule {
    id?: number;
    pattern: string;
    account: string;
    created_at?: string;
    updated_at?: string;
  }

  let rules: Rule[] = [];
  let loading = true;
  let newRule: Rule = { pattern: "", account: "" };
  let editingRule: Rule | null = null;
  let showAddForm = false;
  let showEditForm = false;
  let testPattern = "";
  let testPayee = "";
  let testDescription = "";
  let testAccount = "";
  let testResult = "";
  let testLoading = false;

  onMount(async () => {
    await fetchRules();
  });

  async function fetchRules() {
    loading = true;
    try {
      const response = await fetch("/api/rules");
      const data = await response.json();
      rules = data.rules || [];
    } catch (error) {
      console.error("Error fetching rules:", error);
      toast("Error fetching rules", "danger");
    } finally {
      loading = false;
    }
  }

  async function addRule() {
    try {
      const response = await fetch("/api/rules", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(newRule)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to add rule");
      }

      await fetchRules();
      newRule = { pattern: "", account: "" };
      showAddForm = false;
      toast("Rule added successfully", "success");
    } catch (error) {
      console.error("Error adding rule:", error);
      toast(error.message || "Error adding rule", "danger");
    }
  }

  async function updateRule() {
    if (!editingRule) return;

    try {
      const response = await fetch(`/api/rules/${editingRule.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(editingRule)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to update rule");
      }

      await fetchRules();
      editingRule = null;
      showEditForm = false;
      toast("Rule updated successfully", "success");
    } catch (error) {
      console.error("Error updating rule:", error);
      toast(error.message || "Error updating rule", "danger");
    }
  }

  async function deleteRule(id: number) {
    if (!confirm("Are you sure you want to delete this rule?")) return;

    try {
      const response = await fetch(`/api/rules/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to delete rule");
      }

      await fetchRules();
      toast("Rule deleted successfully", "success");
    } catch (error) {
      console.error("Error deleting rule:", error);
      toast(error.message || "Error deleting rule", "danger");
    }
  }

  function startEdit(rule: Rule) {
    editingRule = { ...rule };
    showEditForm = true;
  }

  async function testRulePattern() {
    testLoading = true;
    testResult = "";

    try {
      const response = await fetch("/api/predict/rules", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          payee: testPayee,
          description: testDescription,
          account: testAccount
        })
      });

      if (response.ok) {
        const data = await response.json();
        testResult = `Matched account: ${data.account}`;
      } else if (response.status === 404) {
        testResult = "No matching rule found";
      } else {
        const error = await response.json();
        throw new Error(error.error || "Test failed");
      }
    } catch (error) {
      console.error("Error testing rule:", error);
      testResult = `Error: ${error.message || "Test failed"}`;
    } finally {
      testLoading = false;
    }
  }

  function formatDate(dateString: string) {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleString();
  }
</script>

<div class="container">
  <div class="box">
    <h1 class="title is-4">Account Prediction Rules</h1>
    <p class="subtitle is-6 mb-4">
      Create regex patterns to automatically assign accounts to transactions based on their details.
    </p>

    <div class="buttons mb-4">
      <button class="button is-primary" on:click={() => (showAddForm = true)}>
        <span class="icon">
          <i class="fas fa-plus"></i>
        </span>
        <span>Add New Rule</span>
      </button>
    </div>

    {#if showAddForm}
      <div class="box">
        <h2 class="title is-5">Add New Rule</h2>
        <div class="field">
          <label class="label">Pattern (Regex)</label>
          <div class="control">
            <input
              class="input"
              type="text"
              placeholder="e.g. (account=Expenses:Unknown AND account=Liabilities:CreditCard:HDFC) AND (payee=~/google/i)"
              bind:value={newRule.pattern}
            />
          </div>
          <p class="help">
            Use regex patterns to match transaction details. Use 'payee=', 'description=', and 'account=' to match specific fields.
          </p>
        </div>

        <div class="field">
          <label class="label">Target Account</label>
          <div class="control">
            <input
              class="input"
              type="text"
              placeholder="e.g. Expenses:Business:Company"
              bind:value={newRule.account}
            />
          </div>
          <p class="help">The account to assign when this rule matches.</p>
        </div>

        <div class="field is-grouped mt-4">
          <div class="control">
            <button class="button is-primary" on:click={addRule}>Save Rule</button>
          </div>
          <div class="control">
            <button class="button is-light" on:click={() => (showAddForm = false)}>Cancel</button>
          </div>
        </div>
      </div>
    {/if}

    {#if showEditForm && editingRule}
      <div class="box">
        <h2 class="title is-5">Edit Rule</h2>
        <div class="field">
          <label class="label">Pattern (Regex)</label>
          <div class="control">
            <input
              class="input"
              type="text"
              placeholder="e.g. (account=Expenses:Unknown AND account=Liabilities:CreditCard:HDFC) AND (payee=~/google/i)"
              bind:value={editingRule.pattern}
            />
          </div>
        </div>

        <div class="field">
          <label class="label">Target Account</label>
          <div class="control">
            <input
              class="input"
              type="text"
              placeholder="e.g. Expenses:Business:Company"
              bind:value={editingRule.account}
            />
          </div>
        </div>

        <div class="field is-grouped mt-4">
          <div class="control">
            <button class="button is-primary" on:click={updateRule}>Update Rule</button>
          </div>
          <div class="control">
            <button class="button is-light" on:click={() => (showEditForm = false)}>Cancel</button>
          </div>
        </div>
      </div>
    {/if}

    <div class="box mt-5">
      <h2 class="title is-5">Test Rules</h2>
      <p class="subtitle is-6 mb-4">
        Test your rules against transaction details to see which account would be assigned.
      </p>

      <div class="columns">
        <div class="column">
          <div class="field">
            <label class="label">Payee</label>
            <div class="control">
              <input
                class="input"
                type="text"
                placeholder="e.g. Google Workspace"
                bind:value={testPayee}
              />
            </div>
          </div>
        </div>
        <div class="column">
          <div class="field">
            <label class="label">Description</label>
            <div class="control">
              <input
                class="input"
                type="text"
                placeholder="e.g. Monthly subscription"
                bind:value={testDescription}
              />
            </div>
          </div>
        </div>
        <div class="column">
          <div class="field">
            <label class="label">Account</label>
            <div class="control">
              <input
                class="input"
                type="text"
                placeholder="e.g. Liabilities:CreditCard:HDFC"
                bind:value={testAccount}
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field">
        <div class="control">
          <button
            class="button is-info"
            on:click={testRulePattern}
            disabled={testLoading}
          >
            <span class="icon">
              <i class="fas fa-flask"></i>
            </span>
            <span>Test Rules</span>
          </button>
        </div>
      </div>

      {#if testResult}
        <div class="notification {testResult.includes('Error') ? 'is-danger' : testResult.includes('No matching') ? 'is-warning' : 'is-success'}">
          {testResult}
        </div>
      {/if}
    </div>

    <div class="mt-5">
      <h2 class="title is-5">Existing Rules</h2>
      {#if loading}
        <div class="has-text-centered">
          <span class="icon is-large">
            <i class="fas fa-spinner fa-pulse"></i>
          </span>
          <p>Loading rules...</p>
        </div>
      {:else if rules.length === 0}
        <div class="notification is-info">
          No rules have been created yet. Add your first rule using the form above.
        </div>
      {:else}
        <div class="table-container">
          <table class="table is-fullwidth is-striped">
            <thead>
              <tr>
                <th>ID</th>
                <th>Pattern</th>
                <th>Target Account</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {#each rules as rule}
                <tr>
                  <td>{rule.id}</td>
                  <td>
                    <div class="is-family-monospace" style="max-width: 400px; overflow-wrap: break-word;">
                      {$obscure ? "********" : rule.pattern}
                    </div>
                  </td>
                  <td>{$obscure ? "********" : rule.account}</td>
                  <td>{formatDate(rule.created_at)}</td>
                  <td>
                    <div class="buttons are-small">
                      <button class="button is-info" on:click={() => startEdit(rule)}>
                        <span class="icon">
                          <i class="fas fa-edit"></i>
                        </span>
                      </button>
                      <button class="button is-danger" on:click={() => deleteRule(rule.id)}>
                        <span class="icon">
                          <i class="fas fa-trash"></i>
                        </span>
                      </button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .container {
    max-width: 1200px;
    margin: 0 auto;
  }
</style>