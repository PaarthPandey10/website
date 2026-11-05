---
layout: page
title: React SDK Live Demo
---
<!-- markdownlint-disable MD033 -->

[`@gemara/react`](https://github.com/gemaraproj/gemara-react) provides headless, server-component-safe React renderers and interactive primitives for [Gemara](https://github.com/gemaraproj/gemara) artifacts. Every renderer emits semantic HTML with stable `data-gemara-*` attributes so applications can bring their own styling.

Use the interactive sandbox below to inspect live Gemara artifacts rendered with `@gemara/react`.

<div class="demo-iframe-container" style="margin: 1.5rem 0; border: 2px solid var(--color-border); border-radius: var(--radius-md); overflow: hidden; background: var(--color-bg-secondary); box-shadow: var(--shadow-sm);">
  <iframe
    id="gemara-demo-frame"
    src="https://gemaraproj.github.io/gemara-react/"
    title="Gemara React Live Artifact Renderer Demo"
    sandbox="allow-scripts"
    loading="lazy"
    referrerpolicy="no-referrer"
    style="width: 100%; min-height: 680px; border: 0; display: block;"
  ></iframe>
</div>

<p class="demo-fallback">
  <a href="https://gemaraproj.github.io/gemara-react/" target="_blank" rel="noopener noreferrer">Open standalone demo in a new tab →</a>
  · <a href="https://github.com/gemaraproj/gemara-react/tree/main/examples/astro" target="_blank" rel="noopener noreferrer">View the Astro demo source</a>
</p>

<noscript>
  <blockquote>
    <strong>Note:</strong> JavaScript is disabled in your browser. View the raw YAML fixtures directly in the <a href="../tutorials/">Tutorials</a> section or explore the <a href="https://github.com/gemaraproj/gemara-react">@gemara/react repository</a>.
  </blockquote>
</noscript>

<script>
(function () {
  var iframe = document.getElementById("gemara-demo-frame");
  if (!iframe) return;

  var isLocalhost = window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1";
  var hostedOrigin = "https://gemaraproj.github.io";
  var localOrigin = "http://127.0.0.1:4321";
  var demoOrigin = new URLSearchParams(window.location.search).get("demo_origin");
  var allowedOrigins = [hostedOrigin];
  if (isLocalhost) {
    allowedOrigins.push(localOrigin);
    if (demoOrigin === localOrigin) iframe.src = demoOrigin + "/";
  }

  function syncTheme() {
    var theme = document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
    if (iframe.contentWindow) {
      iframe.contentWindow.postMessage({ type: "gemara-theme", theme: theme }, "*");
    }
  }

  iframe.addEventListener("load", syncTheme);

  window.addEventListener("message", function (event) {
    if (!event.data || typeof event.data !== "object") return;
    if (event.source !== iframe.contentWindow) return;
    if (event.data.type === "gemara-ready") {
      syncTheme();
    } else if (event.data.type === "gemara-resize" && typeof event.data.height === "number") {
      iframe.style.height = Math.max(480, event.data.height + 16) + "px";
    }
  });

  var observer = new MutationObserver(syncTheme);
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["data-theme"]
  });
})();
</script>

## What This Demo Showcases

- **Headless Artifact Renderers:** Renders Gemara artifacts with `@gemara/react`.
- **Interactive Primitives (`@gemara/react/interactive`):**
  - `FormatTabs` switches between the live rendered preview and the raw YAML source using the WAI-ARIA tabs pattern.
  - `CollapsibleGroup` toggles collapsible control groups on `ControlCatalog` artifacts.
- **Strict Sandbox Isolation:** Runs inside an isolated `<iframe sandbox="allow-scripts">` hosted by the React SDK project.

## Quick Start

Install `@gemara/react` from npm:

```bash
npm install @gemara/react
```

Render a `ControlCatalog` inside your React or Astro application:

```tsx
import { ControlCatalog } from "@gemara/react/control-catalog";
import { GemaraProvider } from "@gemara/react/provider";
import { isControlCatalog } from "@gemara/react";

export function CatalogView({ data }: { data: unknown }) {
  if (!isControlCatalog(data)) {
    throw new Error("Invalid ControlCatalog artifact");
  }

  return (
    <GemaraProvider>
      <ControlCatalog data={data} headingLevel={2} />
    </GemaraProvider>
  );
}
```

## Related Resources

- **[@gemara/react Repository](https://github.com/gemaraproj/gemara-react)** — Source code, `data-gemara-*` styling contract, and API documentation
- **[SDK Overview](./index.html)** — Versioning and architecture across Gemara SDKs
- **[Tutorials](../tutorials/)** — Step-by-step guides for the YAML artifacts rendered in this demo
