---
title: "Preservation Artifacts"
date: 2024-04-01
weight: 10
---

# Velociraptor Triage Artifacts

This repository contains the compiler used to assemble and manage a
number of Velociraptor artifacts for collecting files from the
endppoint for the purpose of preservation and triage.

## What are preservation artifacts?

In an incident we need to rapidly triage endpoints to establish which
hosts were forensically relevant for the investigation. This may
involve collecting Velociraptor artifacts such as
`Windows.Hayabusa.Rules` or other VQL artifacts that parse and analyze
the forensic artifacts on the endpoint (e.g. apply time boxing, hunt
for common IOCs etc).

While Velociraptor excels in this rapid triaging approach, once we
identify the relevant hosts, we often want to preserve the state of
these endpoints as much as possible - both for evidentiary
preservation and to facilitate future analysis.

This might be because the endpoint is about to be destroyed or
re-imaged, or we might want to preserve critical digital forensic
evidence on the endpoint.

This `Preservtion` or `Acquisition` phase task usually involves
capturing raw files from the endpoint.

### Historical development

Velociraptor's `Windows.KapeFiles.Targets` artifact is based on the
[KapeFiles](https://github.com/EricZimmerman/KapeFiles) project. This
project collects community contributed `Kape Targets` that specify and
curate typical files relevant to the typical forensic investigation.

Since a complete bit for bit acquisition of the endpoint is no longer
practical (with typical disk sized being too large to copy in a
reasonable time), it becomes important to collect only relevant files
from the endpoint. This type of acquisition is called `Triage
Acquisition` - as in to say we only acquire some files to enable us to
quickly determine if the endpoint is relevant.

When selectively collecting files from the endpoint we need to make
compromises and choose which files are relevant to the current
investigation.

The `KapeFiles` project defines a `Target` which identifies a set of
files relevant to a particular collection. For example, the
`BraveBrowser` target lists a set of glob expressions relevant to all
the different artifacts used by the Brave Browser.

This makes it easier to collect these relevant files - one simply
needs to collect the target `BraveBrowser` to be sure to check all
locations where the browser keeps relevant files.

The nice thing about the `KapeFiles` Targets is that each target may
refer to other targets. So the analyst can create higher level targets
which focus on generic collection types.

For example, the `WebBrowsers` target specifies multiple `References`
to other targets. When the user selects to collect this `WebBrowsers`
target, Velociraptor will expand the target to also include the
`BraveBrowser` target as well and other targets like `Chrome` etc.

This allows the user to select the high level target in order to
automatically include the low level targets easily.

At the highest level the `SANS_Triage` or `KapeTriage` targets select
many sub-targets and can be used to trigger a reasonable triage
collection very easily.

## The format of a collection target

This project aims to build and extend on the `KapeFiles` project's
definition of a `Target`.

### The Target file format

A target is specified in a YAML file (with the extension `.yaml` or
`.tkape`) with the following fields:

```yaml
Name: The name of the target (If not set taken from the filename)
Description: What the target represents
Author: The author.
Rules:
- Name: Name of the rule
  Category: The category of the rule
  Comment: Some description of this rule.
  Glob: A glob expression to find the files related to the rule.
  Ref: The name of a target to refer to instead.
```

The target specifies a list of rules; Each rule can specify either a
`Glob` expression to indicate which files to capture, or the name of
another target reference.

## Triage Collection Artifacts

You can view the artifacts managed by this project on the sidebar to
the left:

* [Windows.KapeFiles.Targets]({{< ref "docs/Windows.KapeFiles.Targets/rules/" >}}) is the original glob based Windows file collector.
