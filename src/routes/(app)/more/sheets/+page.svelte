<script lang="ts">
  import { goto } from "$app/navigation";
  import FileModal from "$lib/components/FileModal.svelte";
  import { ajax } from "$lib/utils";
  import * as toast from "bulma-toast";

  let modalOpen = false;
  function openCreateModal() {
    modalOpen = true;
  }

  async function createFile(destinationFile: string) {
    destinationFile = destinationFile.trim() + ".paisa";
    const { saved, message } = await ajax("/api/sheets/save", {
      method: "POST",
      body: JSON.stringify({ name: destinationFile, content: "", operation: "create" }),
      background: true
    });

    if (saved) {
      toast.toast({
        message: `Created <b><a href="/more/sheets/${encodeURIComponent(
          destinationFile
        )}">${destinationFile}</a></b>`,
        type: "is-success",
        duration: 5000
      });

      await goto(`/more/sheets/${encodeURIComponent(destinationFile)}`);
    } else {
      toast.toast({
        message: `Failed to create ${destinationFile}. reason: ${message}`,
        type: "is-danger",
        duration: 10000
      });
    }
  }
</script>

<FileModal
  bind:open={modalOpen}
  on:save={(e) => createFile(e.detail)}
  label="Create"
  placeholder="scratch"
  help="Filename without any extension"
/>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-6 mx-auto">
        <div class="flex items-center justify-center mt-5">
          <div class="field">
            <p class="control">
              <button class="button is-medium is-link" on:click={(_e) => openCreateModal()}>
                <span class="icon is-small">
                  <i class="fas fa-file-circle-plus" />
                </span>
                <span>Create</span>
              </button>
            </p>
            <p class="mt-2 has-text-grey has-text-bold">Create your first sheet</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
