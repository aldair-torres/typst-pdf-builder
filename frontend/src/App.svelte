<script>
  import { OpenDirectory, ScanProject, Build } from '../wailsjs/go/main/App'

  let workDir = $state('')
  let scanning = $state(false)

  let rootTypes = $state([])
  let langFolders = $state([])
  let langTypes = $state({})

  let buildType = $state('')
  let selectedLangs = $state([])
  let selectedRootTyp = $state('')
  let selectedLangTyps = $state({})

  let cols = $state('1')
  let media = $state('digital')
  let production = $state('false')
  let audience = $state('')
  let coverImage = $state('')
  let product = $state('')
  let publication = $state('')
  let productLine2 = $state('')
  let publicationLine2 = $state('')
  let extraArgs = $state('')

  let building = $state(false)
  let buildResult = $state(null)
  let showModal = $state(false)
  let showInvalidModal = $state(false)

  let folderSelected = $derived(workDir !== '')
  let hasLangs = $derived(langFolders.length > 0)
  let hasRoot = $derived(rootTypes.length > 0)
  let formDisabled = $derived(!folderSelected || scanning)

  $effect(() => {
    if (rootTypes.length === 1) selectedRootTyp = rootTypes[0]
  })

  let canCompile = $derived(
    folderSelected && buildType !== '' && (
      buildType === 'languages'
        ? selectedLangs.length > 0
        : selectedRootTyp !== ''
    )
  )

  async function selectFolder() {
    const dir = await OpenDirectory()
    if (!dir) return
    if (dir !== workDir) {
      buildType = ''
      selectedLangs = []
      selectedRootTyp = ''
      selectedLangTyps = {}
      cols = '1'
      media = 'digital'
      production = 'false'
      audience = ''
      coverImage = ''
      product = ''
      publication = ''
      productLine2 = ''
      publicationLine2 = ''
      extraArgs = ''
      buildResult = null
      showInvalidModal = false
      workDir = dir
    }
    scanning = true
    const result = await ScanProject(dir)
    rootTypes = result.rootTypes ?? []
    langFolders = result.langFolders ?? []
    langTypes = result.langTypes ?? {}
    product = result.product ?? ''
    publication = result.publication ?? ''
    for (const lang of langFolders) {
      const typs = langTypes[lang] ?? []
      if (typs.length === 1) selectedLangTyps[lang] = typs[0]
    }
    scanning = false
    if ((result.rootTypes ?? []).length === 0 && (result.langFolders ?? []).length === 0) {
      showInvalidModal = true
    }
  }

  function toggleLang(lang) {
    if (selectedLangs.includes(lang)) {
      selectedLangs = selectedLangs.filter(l => l !== lang)
    } else {
      selectedLangs = [...selectedLangs, lang]
    }
  }

  async function compile() {
    building = true
    try {
      buildResult = await Build({
        mode: buildType,
        langs: selectedLangs,
        langTyps: selectedLangTyps,
        rootTyp: selectedRootTyp,
        cols,
        media,
        audience,
        production: production === 'true',
        coverImage,
        product,
        publication,
        productLine2,
        publicationLine2,
        extraArgs,
      })
    } catch (e) {
      buildResult = { success: false, log: String(e), errors: [String(e)] }
    }
    building = false
    showModal = true
  }
</script>

<div class="min-h-screen bg-base-100">
  <div class="navbar bg-base-200/80 backdrop-blur-sm border-b border-base-300 sticky top-0 z-50 px-6">
    <div class="flex-1">
      <span class="text-xl font-bold tracking-tight">
        <span class="text-primary">Typst</span> PDF Builder
      </span>
    </div>
    <div class="flex-none">
      <button
        class="btn btn-primary"
        onclick={compile}
        disabled={!canCompile || building}
      >
        {#if building}
          <span class="loading loading-spinner loading-sm"></span>
          Compiling…
        {:else}
          Compile
        {/if}
      </button>
    </div>
  </div>

  <div class="max-w-2xl mx-auto px-4 py-6 space-y-5">

    <div class="card bg-base-200 shadow-sm">
      <div class="card-body p-5">
        <h2 class="card-title text-base font-semibold">Working Directory</h2>
        <div class="join w-full">
          <input
            type="text"
            class="input input-bordered join-item w-full text-sm"
            value={workDir || 'No folder selected'}
            readonly
          />
          <button class="btn btn-primary join-item" onclick={selectFolder} disabled={scanning}>
            Browse…
          </button>
        </div>
        {#if scanning}
          <div class="flex items-center gap-3 mt-3">
            <span class="loading loading-spinner text-primary loading-sm"></span>
            <span class="text-sm opacity-70">Scanning project…</span>
          </div>
        {/if}
      </div>
    </div>

    <div class="card bg-base-200 shadow-sm {formDisabled ? 'opacity-60 pointer-events-none' : ''}">
      <div class="card-body p-5">
        <h2 class="card-title text-base font-semibold">What to build</h2>
        <div class="space-y-2">
          {#if hasLangs || !folderSelected}
            <label class="flex items-center gap-3 cursor-pointer">
              <input
                type="radio"
                name="buildType"
                value="languages"
                bind:group={buildType}
                class="radio radio-sm radio-primary"
                disabled={formDisabled}
              />
              <span class="text-sm">Build one or more languages</span>
            </label>
          {/if}
          {#if hasRoot || !folderSelected}
            <label class="flex items-center gap-3 cursor-pointer">
              <input
                type="radio"
                name="buildType"
                value="multi"
                bind:group={buildType}
                class="radio radio-sm radio-primary"
                disabled={formDisabled}
              />
              <span class="text-sm">Build multilingual booklet</span>
            </label>
          {/if}
        </div>

        {#if buildType === 'languages'}
          <div class="mt-3">
            <details class="dropdown w-full">
              <summary class="btn btn-outline btn-sm w-full justify-between font-normal">
                <span class={selectedLangs.length === 0 ? 'opacity-50' : ''}>
                  {selectedLangs.length === 0 ? 'Select languages…' : selectedLangs.join(', ')}
                </span>
                <svg class="w-3 h-3 opacity-50 ml-2 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                </svg>
              </summary>
              <div class="dropdown-content bg-base-100 rounded-box shadow w-full mt-1 z-10 p-2 max-h-60 overflow-y-auto">
                {#each langFolders as lang}
                  <div class="flex items-center gap-2 px-3 py-1.5">
                    <input
                      type="checkbox"
                      checked={selectedLangs.includes(lang)}
                      onchange={() => toggleLang(lang)}
                      class="checkbox checkbox-sm checkbox-primary"
                    />
                    <span class="text-sm flex-1">{lang}</span>
                    {#if (langTypes[lang] ?? []).length > 1}
                      <select bind:value={selectedLangTyps[lang]} class="select select-bordered select-xs w-auto">
                        {#each langTypes[lang] as typ}
                          <option value={typ}>{typ}</option>
                        {/each}
                      </select>
                    {/if}
                  </div>
                {/each}
              </div>
            </details>
          </div>
        {/if}

        {#if buildType === 'multi' && rootTypes.length > 1}
          <div class="mt-3">
            <details class="dropdown w-full">
              <summary class="btn btn-outline btn-sm w-full justify-between font-normal">
                <span class={!selectedRootTyp ? 'opacity-50' : ''}>
                  {selectedRootTyp || 'Select booklet…'}
                </span>
                <svg class="w-3 h-3 opacity-50 ml-2 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                </svg>
              </summary>
              <div class="dropdown-content bg-base-100 rounded-box shadow w-full mt-1 z-10 p-2">
                {#each rootTypes as typ}
                  <label class="flex items-center gap-2 px-3 py-1.5 cursor-pointer">
                    <input
                      type="radio"
                      name="rootTyp"
                      value={typ}
                      bind:group={selectedRootTyp}
                      class="radio radio-sm radio-primary"
                    />
                    <span class="text-sm">{typ}</span>
                  </label>
                {/each}
              </div>
            </details>
          </div>
        {/if}
      </div>
    </div>

    <div class="card bg-base-200 shadow-sm {formDisabled ? 'opacity-60 pointer-events-none' : ''}">
      <div class="card-body p-5">
        <h2 class="card-title text-base font-semibold">Cover Text</h2>
        <div class="grid grid-cols-2 gap-4">
          <label class="form-control">
            <div class="label py-1"><span class="label-text text-xs opacity-70">Product</span></div>
            <input
              type="text"
              bind:value={product}
              placeholder="e.g. Product name"
              class="input input-bordered w-full text-sm"
              disabled={formDisabled}
            />
          </label>
          <label class="form-control">
            <div class="label py-1"><span class="label-text text-xs opacity-70">Product line 2</span></div>
            <input
              type="text"
              bind:value={productLine2}
              placeholder="optional"
              class="input input-bordered w-full text-sm"
              disabled={formDisabled}
            />
          </label>
        </div>
        <div class="grid grid-cols-2 gap-4 mt-2">
          <label class="form-control">
            <div class="label py-1"><span class="label-text text-xs opacity-70">Publication</span></div>
            <input
              type="text"
              bind:value={publication}
              placeholder="e.g. Publication name"
              class="input input-bordered w-full text-sm"
              disabled={formDisabled}
            />
          </label>
          <label class="form-control">
            <div class="label py-1"><span class="label-text text-xs opacity-70">Publication line 2</span></div>
            <input
              type="text"
              bind:value={publicationLine2}
              placeholder="optional"
              class="input input-bordered w-full text-sm"
              disabled={formDisabled}
            />
          </label>
        </div>
      </div>
    </div>

    <div class="card bg-base-200 shadow-sm {formDisabled ? 'opacity-60 pointer-events-none' : ''}">
      <div class="card-body p-5">
        <h2 class="card-title text-base font-semibold">Build Options</h2>

        <div class="grid grid-cols-2 gap-6">
          <div class="space-y-4">
            {#if buildType !== 'multi'}
              <div>
                <h3 class="text-xs font-semibold opacity-50 uppercase tracking-wide mb-2">Columns</h3>
                <div class="flex gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="cols"
                      value="1"
                      bind:group={cols}
                      class="radio radio-sm radio-primary"
                      disabled={formDisabled}
                    />
                    <span class="text-sm">1 column</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="cols"
                      value="2"
                      bind:group={cols}
                      class="radio radio-sm radio-primary"
                      disabled={formDisabled}
                    />
                    <span class="text-sm">2 columns</span>
                  </label>
                </div>
              </div>
            {/if}

            <div>
              <h3 class="text-xs font-semibold opacity-50 uppercase tracking-wide mb-2">Media</h3>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="media"
                    value="digital"
                    bind:group={media}
                    class="radio radio-sm radio-primary"
                    disabled={formDisabled}
                  />
                  <span class="text-sm">Digital</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="media"
                    value="printed"
                    bind:group={media}
                    class="radio radio-sm radio-primary"
                    disabled={formDisabled}
                  />
                  <span class="text-sm">Printed</span>
                </label>
              </div>
            </div>

            <div>
              <h3 class="text-xs font-semibold opacity-50 uppercase tracking-wide mb-2">Production Mode</h3>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="production"
                    value="false"
                    bind:group={production}
                    class="radio radio-sm radio-primary"
                    disabled={formDisabled}
                  />
                  <span class="text-sm">No</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="production"
                    value="true"
                    bind:group={production}
                    class="radio radio-sm radio-primary"
                    disabled={formDisabled}
                  />
                  <span class="text-sm">Yes</span>
                </label>
              </div>
            </div>
          </div>

          <div class="space-y-3">
            <label class="form-control">
              <div class="label py-1"><span class="label-text text-xs opacity-70">Audience</span></div>
              <input
                type="text"
                bind:value={audience}
                placeholder="optional"
                class="input input-bordered w-full text-sm"
                disabled={formDisabled}
              />
            </label>
            <label class="form-control">
              <div class="label py-1"><span class="label-text text-xs opacity-70">Cover Image</span></div>
              <input
                type="text"
                bind:value={coverImage}
                placeholder="Path to cover image (optional)"
                class="input input-bordered w-full text-sm"
                disabled={formDisabled}
              />
            </label>
            <label class="form-control">
              <div class="label py-1"><span class="label-text text-xs opacity-70">Extra Typst Arguments</span></div>
              <input
                type="text"
                bind:value={extraArgs}
                placeholder="optional"
                class="input input-bordered w-full text-sm"
                disabled={formDisabled}
              />
            </label>
          </div>
        </div>


      </div>
    </div>

  </div>
</div>

{#if showInvalidModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="text-lg font-bold">No Valid Typst Files Found</h3>
      <p class="py-4 text-sm opacity-80">
        The folder you selected doesn't contain valid
        <code class="badge badge-ghost badge-sm">.typ</code>
        files. Choose a valid project folder and try again.
      </p>
      <div class="modal-action">
        <button class="btn btn-primary btn-sm" onclick={() => { showInvalidModal = false }}>OK</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button onclick={() => { showInvalidModal = false }}>close</button>
    </form>
  </div>
{/if}

{#if showModal && buildResult}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <h3 class="text-lg font-bold">
        {#if buildResult.success}
          <span class="text-success">Build Successful</span>
        {:else}
          <span class="text-error">Build Failed</span>
        {/if}
      </h3>
      <pre class="bg-base-300 rounded-box p-4 text-xs leading-relaxed max-h-80 overflow-auto whitespace-pre-wrap font-mono mt-3">{buildResult.log || '(no output)'}</pre>
      {#if buildResult.errors && buildResult.errors.length > 0}
        <div class="mt-3 space-y-1.5">
          {#each buildResult.errors as err}
            <div class="text-error text-sm">{err}</div>
          {/each}
        </div>
      {/if}
      <div class="modal-action">
        <button class="btn btn-primary btn-sm" onclick={() => { showModal = false }}>Close</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button onclick={() => { showModal = false }}>close</button>
    </form>
  </div>
{/if}
