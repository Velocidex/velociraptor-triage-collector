---
title: "Developing Targets"
date: 2025-08-11
weight: 20
bookToc: false
IconClass: fa-solid fa-code
---

# Developing targets

The following describes how to add new custom targets to the artifact.

For the purpose of this exercise, I will build the artifact on a Linux
machine, and test it on a Windows VM. There is a shared directory
between these systems - `/shared/`, mapped to the `F:` drive on
Windows.

## Using the notebook on the target system

Create a new notebook where you can test the target.

To rebuild the artifact on the Linux system:

```bash
$ make compiler verify
```

This will build the artifact and run Velociraptor's artifact
verification on the output. The artifact verifier will flag many
issues with the VQL which can arise by using some more complex rules.

I will now copy the artifacts to the shared folder

```bash
$ cp output/*.yaml /shared/artifacts/
```

In the test VM, I can import the updated version of the artifact using
the following VQL query in a notebook cell.

```sql
SELECT artifact_set(definition=read_file(filename=OSPath)).name AS Name, OSPath
FROM glob(globs="F:/artifacts/*.yaml")
```

I can then collect it inside the notebook cell

```sql
SELECT * FROM Artifact.Windows.Triage.Targets(HighLevelTargets="[\"_BasicCollection\"]")
```

Alternatively I can collect it with the command line

```text
velociraptor-v0.75.5-windows-amd64.exe  --definitions f:\artifacts\ artifacts collect Windows.Triage.Targets --args HighLevelTargets=[\"_BasicCollection\"] -v --output test.zip
```

Note that the high level targets are encoded as a json array.
