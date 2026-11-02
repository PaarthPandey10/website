---
layout: page
title: SDKs
---

Language-specific SDKs provide programmatic access to Gemara documents. SDK types are generated from CUE schemas, ensuring consistency between validation and programmatic access.

## Available SDKs

- [Go SDK](./go-sdk.html) — Type-safe Go APIs for reading, writing, and manipulating Gemara documents
- [Python SDK](./python-sdk.html) — Typed Pydantic v2 models and loader utilities for Gemara documents
- [React SDK](./react-sdk.html) — Headless React renderers and UI components for Gemara documents

## Versioning and Maintenance

Project release deliverables are divided into the **Core Specification** and language-specific **SDKs**.

* **Core Specification Release:** The CUE schemas are versioned and released as the core specification. These schemas implement the Model, which is published separately.
* **SDK Releases:** Language-specific implementations that support programmatic access to Gemara documents. SDK types are generated from the CUE schemas.

Each maintains its own independent [SemVer](https://semver.org/) lifecycle.

### SDK Release Versioning

SDK releases are versioned independently to allow for rapid iteration on tooling, bug fixes, and utilities without requiring core specification version bumps.

* The CUE schemas in the core specification release are the source of truth for validation.
* SDKs must explicitly document which core specification release version they support.
* When a new core specification release is published, SDKs regenerate their types from the updated schemas and release new versions that support the updated specification.

## Relationship to Other Components

### [The Model](../model)
Provides the conceptual foundation. SDK types correspond to elements in the model.

### [The Schemas](../schema/)
SDK types are generated from the CUE schemas. The schemas are the source of truth for data structures.

### [The Specifications](../specifications/)
Tool and resolution specifications define the normative operational algorithms that SDKs and tools implement for downstream consumers.
