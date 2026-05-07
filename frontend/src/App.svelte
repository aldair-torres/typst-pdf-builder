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
  let extraArgs = $state('')

  let building = $state(false)
  let buildResult = $state(null)
  let showModal = $state(false)
  let showInvalidModal = $state(false)

  let folderSelected = $derived(workDir !== '')
  let hasLangs = $derived(langFolders.length > 0)
  let hasRoot = $derived(rootTypes.length > 0)

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
        extraArgs,
      })
    } catch (e) {
      buildResult = { success: false, log: String(e), errors: [String(e)] }
    }
    building = false
    showModal = true
  }
</script>

<main class="container-fluid">
  <nav>
    <ul>
      <li><h4>Typst PDF builder</h4></li>
    </ul>
  </nav>
  <!-- Working Directory -->
  <article>
    <header><strong>Working Directory</strong></header>
    <div role="group">
      <input type="text" value={workDir || 'No folder selected'} readonly />
      <button onclick={selectFolder}>Browse&hellip;</button>
    </div>
    {#if scanning}
      <p aria-busy="true">Scanning project&hellip;</p>
    {/if}
  </article>

  <form>
    <!-- What to build -->
    <fieldset disabled={!folderSelected || scanning}>
      <legend><strong>What to build</strong></legend>
      {#if hasLangs || !folderSelected}
      <label>
        <input type="radio" name="buildType" value="languages" bind:group={buildType} />
        Build one or more languages
      </label>
      {/if}
      {#if hasRoot || !folderSelected}
      <label>
        <input type="radio" name="buildType" value="multi" bind:group={buildType} />
        Build multilingual booklet
      </label>
      {/if}
    </fieldset>

    <!-- Language selection (languages mode) -->
    {#if buildType === 'languages'}
    <details class="dropdown">
      <summary>
        {selectedLangs.length === 0 ? 'Select languages…' : selectedLangs.join(', ')}
      </summary>
      <ul>
        {#each langFolders as lang}
        <li>
          <label>
            <input
              type="checkbox"
              checked={selectedLangs.includes(lang)}
              onchange={() => toggleLang(lang)}
            />
            {lang}
            {#if (langTypes[lang] ?? []).length > 1}
              &nbsp;
              <select bind:value={selectedLangTyps[lang]} style="display:inline-block;width:auto">
                {#each langTypes[lang] as typ}
                  <option value={typ}>{typ}</option>
                {/each}
              </select>
            {/if}
          </label>
        </li>
        {/each}
      </ul>
    </details>
    {/if}

    <!-- Booklet selection (multi mode, only when >1 root typ) -->
    {#if buildType === 'multi' && rootTypes.length > 1}
    <details class="dropdown">
      <summary>
        {selectedRootTyp || 'Select booklet…'}
      </summary>
      <ul>
        {#each rootTypes as typ}
        <li>
          <label>
            <input type="radio" name="rootTyp" value={typ} bind:group={selectedRootTyp} />
            {typ}
          </label>
        </li>
        {/each}
      </ul>
    </details>
    {/if}

    <!-- Build options -->
    <fieldset disabled={!folderSelected || scanning}>
      <legend><strong>Build Options</strong></legend>

      <div class="grid">
        <div>
          {#if buildType !== 'multi'}
          <fieldset>
            <legend>Columns</legend>
            <label>
              <input type="radio" name="cols" value="1" bind:group={cols} />
              1 column
            </label>
            <label>
              <input type="radio" name="cols" value="2" bind:group={cols} />
              2 columns
            </label>
          </fieldset>
          {/if}

          <fieldset>
            <legend>Media</legend>
            <label>
              <input type="radio" name="media" value="digital" bind:group={media} />
              Digital
            </label>
            <label>
              <input type="radio" name="media" value="printed" bind:group={media} />
              Printed
            </label>
          </fieldset>

          <fieldset>
            <legend>Production Mode</legend>
            <label>
              <input type="radio" name="production" value="false" bind:group={production} />
              No
            </label>
            <label>
              <input type="radio" name="production" value="true" bind:group={production} />
              Yes
            </label>
          </fieldset>
        </div>

        <div>
          <label>
            Audience
            <input type="text" bind:value={audience} placeholder="optional" />
          </label>

          <label>
            Cover Image
            <input type="text" bind:value={coverImage} placeholder="Path to cover image (optional)" />
          </label>

          <label>
            Extra Typst Arguments
            <input type="text" bind:value={extraArgs} placeholder="optional" />
          </label>
        </div>
      </div>

      <button onclick={compile} disabled={!canCompile || building} aria-busy={building}>
        {building ? 'Compiling…' : 'Compile'}
      </button>
    </fieldset>
  </form>
</main>

<!-- Invalid folder modal -->
{#if showInvalidModal}
<dialog open>
  <article>
    <h4>⚠️ No valid Typst files found</h4>
    <p>The folder you selected doesn't contain valid <code>.typ</code> files. Choose a valid project folder and try again.</p>
    <footer>
      <button onclick={() => { showInvalidModal = false }}>OK</button>
    </footer>
  </article>
</dialog>
{/if}

<!-- Result modal -->
{#if showModal && buildResult}
<dialog open>
  <article>
    <h4>{buildResult.success ? '✅ Build successful' : '❌ Build failed'}</h4>
    <pre style="overflow:auto;max-height:50vh;white-space:pre-wrap">{buildResult.log || '(no output)'}</pre>
    {#if buildResult.errors && buildResult.errors.length > 0}
      {#each buildResult.errors as err}
        <p style="color:var(--pico-color-red-500)">{err}</p>
      {/each}
    {/if}
    <footer>
      <button onclick={() => { showModal = false }}>Close</button>
    </footer>
  </article>
</dialog>
{/if}
