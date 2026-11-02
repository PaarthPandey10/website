---
layout: page
title: The Need for Machine-Optimized Documentation Standards
---

## Going Beyond Machine-Readable

As mentioned in the _[Prior Work](./03-foundational-concepts)_, the *GRC Engineering Model for Automated Risk Assessments* is the product of many previous iterations and learnings. 

Immense value has come from the work on OSCAL, where a large body of machine-readable documents have been crafted both publicly and privately by highly regulated organizations. The work done in the OSCAL space cannot be overlooked or discounted, as it has provided a great benefit to countless firms.

However, there is something to be said about the benefits of *optimizing* for machines when crafting machine-readable GRC documents. Previous efforts to craft relevant schemas have fallen short in two key areas.

The first, and most obvious, shortcoming comes from tools that misappropriate the term “Policy” to describe security tooling configurations. As useful as these resources may be, far too many organizations waste the potential of highly-trained GRC and security professionals by making them responsible for managing tooling configurations under the pretense of writing Policy-as-Code.

The second shortcoming is more difficult to spot, and it comes in the form of a specification which allows for customized fields.

Although it may sound innocent or even necessary, the addition of customization of GRC document schemas results in the need for custom-tailored tools that have foreknowledge of those fields. The tradeoff for this flexibility is a loss of structural clarity needed to build widely adopted tooling around the data. While a schema should be flexible enough to support a variety of uses, it should also be opinionated enough to allow an automation ecosystem to rise up and thrive around it.

By drawing on the positive lessons and addressing known issues, we can enable organizations to engineer highly efficient GRC ecosystems that are robust enough to address complex problems, but comprehensible enough for any professional to quickly get up to speed, all while maintaining the ability to map to other formats, such as OSCAL.

Achieving an opinionated, standardized schema for each activity type will allow for rapid industry-wide acceleration of automated Risk Assessments.

## Improvements for Humans; Improvements for AI

An organization that adopts an optimized GRC Engineering strategy can gain benefits for both humans and artificial intelligence.

Communication and handoffs are a well-known obstacle for highly regulated firms. Field observations consistently reveal challenges distributing Policies to impacted parties, managing changes, accidental rework, and even months spent on unnecessary work due to misunderstandings.

Firms that standardize document schemas can develop the necessary tools to get documents to the right place at the right time. The database and APIs necessary to build these communication optimizations provide a secondary benefit: they act as the foundation for Model Context Protocol (MCP).

When a properly designed system such as MCP or AI-skills is overlaid on a machine-optimized GRC database, AI capabilities benefit from replacing vague, unstructured prose with unique technical identifiers and explicit resource mappings. This structural clarity allows the model to act upon requirements, minimizing the need for probabilistic inference of intent and leading to more deterministic outcomes. 

When partnered with an Agent2Agent (A2A) enabled system that allows GRC specialized systems to provide precisely contextualized information to engineering specialized systems in a near deterministic manner, it has been observed that AI systems will require dramatically fewer resources, security tools increase in accuracy, less time is spent on false positives or negatives, and more time is spent delivering real value to the organization.

- **< Previous Page**: [Layer 7](./07.4-Layer-7)
- **> Next Page**: [Conclusion](./09-conclusion)
