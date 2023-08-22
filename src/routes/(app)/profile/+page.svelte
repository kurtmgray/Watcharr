<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { createExportUrl } from "@/lib/util/api";
  import { getOrdinalSuffix, monthsShort, toggleTheme } from "@/lib/util/helpers";
  import { appTheme } from "@/store";
  import type { Profile, DownloadFormat } from "@/types";
  import axios from "axios";

  $: selectedTheme = $appTheme;

  let downloadUrl: string | undefined;
  let format: DownloadFormat;

  async function getProfile() {
    return (await axios.get(`/profile`)).data as Profile;
  }

  function formatDate(d: Date) {
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${
      monthsShort[d.getMonth()]
    } ${d.getFullYear()}`;
  }

  function getFileExtension(format: DownloadFormat) {
  switch (format) {
    case 'json':
      return 'txt';
    case 'xml':
      return 'xml';
    case 'csv':
      return 'csv';
    default:
      return 'txt';
  }
}

  async function createDownloadLink(e: MouseEvent) {
    const button = e.target as HTMLButtonElement;
    format = button.dataset.format as DownloadFormat;
    downloadUrl = createExportUrl(format);
    
  }

</script>

<div class="content">
  <div class="inner">
    <h2>Hey {localStorage.getItem("username")}</h2>

    <div class="stats">
      {#await getProfile()}
        <Spinner />
      {:then profile}
        <div>
          <span>{formatDate(new Date(profile.joined))}</span>
          <span>Joined</span>
        </div>
        <div>
          <span class="large">{profile.moviesWatched}</span>
          <span>Movies Watched</span>
        </div>
        <div>
          <span class="large">{profile.showsWatched}</span>
          <span>Shows Watched</span>
        </div>
      {:catch err}
        <Error error={err} pretty="Failed to get stats!" />
      {/await}
    </div>

    <div class="settings">
      <h3 class="norm">Settings</h3>

      <h4 class="norm">Theme</h4>
      <div class="theme">
        <button
          class={`plain${selectedTheme === "light" ? " selected" : ""}`}
          id="light"
          on:click={() => toggleTheme("light")}
        >
          light
        </button>
        <button
          class={`plain${selectedTheme === "dark" ? " selected" : ""}`}
          id="dark"
          on:click={() => toggleTheme("dark")}
        >
          dark
        </button>
      </div>

      <button data-format="json" on:click={createDownloadLink}>Export JSON</button>
      <button data-format="csv" on:click={createDownloadLink}>Export CSV</button>
      <button data-format="xml" on:click={createDownloadLink}>Export XML</button>
      
      {#if downloadUrl}
        <a href={downloadUrl} download={`data.${getFileExtension(format)}`}>
          <button on:click={() => {downloadUrl = undefined}}>
            Download {format.toUpperCase()}
          </button>
        </a>
      {/if}
    </div>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .inner {
      min-width: 400px;
      max-width: 400px;

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }

      @media screen and (max-width: 440px) {
        width: 100%;
        min-width: unset;
      }
    }
  }

  .stats {
    display: flex;
    flex-flow: row;
    gap: 12px;
    margin-top: 15px;

    @media screen and (max-width: 440px) {
      flex-wrap: wrap;
    }

    > div {
      display: flex;
      flex-flow: column;
      flex-grow: 1;
      padding: 20px 15px;
      background-color: $accent-color;
      border-radius: 8px;

      > span:first-child {
        font-weight: bold;
        font-size: 20px;

        &.large {
          font-size: 32px;
        }
      }

      > span:last-child {
        margin-top: auto;
      }
    }
  }

  .settings {
    display: flex;
    flex-flow: column;
    width: 100%;

    h3 {
      margin-bottom: 15px;
      font-variant: small-caps;
    }

    h4 {
      margin-bottom: 0px;
      margin-left: 15px;
    }

    .theme {
      display: flex;
      gap: 10px;
      margin: 20px;
      margin-top: 15px;

      & > button {
        width: 50%;
        height: 80px;
        border-radius: 10px;
        outline: 3px solid;
        font-size: 20px;
        text-transform: uppercase;
        font-family: "Rampart One";
        color: transparent;
        transition: all 200ms ease-in;

        &#light {
          background-color: white;
          outline-color: $accent-color;
          &:hover {
            color: black;
            -webkit-text-stroke: 0.5px black;
          }
        }

        &#dark {
          background-color: black;
          outline-color: white;
          &:hover {
            color: white;
            -webkit-text-stroke: 0.5px white;
          }
        }

        &.selected {
          outline-color: gold !important;
        }
      }
    }
  }
</style>
