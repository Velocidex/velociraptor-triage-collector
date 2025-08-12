---
title: "Windows.KapeFiles.Targets"
date: 2025-08-11
weight: 10
bookToc: false
IconClass: fa-solid fa-desktop
---

# Windows.KapeFiles.Targets

This artifact is built automatically from the
[KapeFiles](https://github.com/EricZimmerman/KapeFiles) project.

You can [download the artifact](Windows.KapeFiles.Targets.zip) for manual import into Velociraptor.

The description below explains how to use this artifact in practice.

The artifact will generate a list of globs and prepend the device name
to each glob. Velociraptor's `glob()` plugin implementation is very
efficient and minimizes the number of passes it needs to make over the
filesystem, when using multiple glob expressions at the same time.

Therefore the artifact first traverses all the rules to build a large
list of glob expressions, which it uses to search for candidate files.

## Parameters

1. **Devices**: This is the list of drives that should be
   considered. By default we only consider the `C:` drive but if you
   might have other drives in use, then we consider those as well. The
   drive name is prepended to each glob specified by the different
   rules to begin searching on that device.

2. **DropVerySlowRules**: Some targets specify globs which need to
   examine every file on the disk. For example,
   `DirectoryTraversal_AudioFiles` has a glob similar to
   `C:\**\*.{3gp,aa,aac,act,aiff}`.

   This type of search is very slow as it needs to examine every file
   on disk. By default we disable these rules because they are too
   slow to be useful. If you really want them enabled, switch this
   setting off, but collection time will increase significantly.

3. **VSS_MAX_AGE_DAYS**: By default we do not consider Volume Shadow
   Copies during file collection. However, if you set this value to a
   number larger than 0, we consider this many days worth of VSS
   copies.

   This setting causes Velociraptor to repeat the search on all VSS
   copies within the specified time limit, and check for changed files
   between VSS copies. If the file has changed (or maybe deleted)
   between the different VSS copies, then Velociraptor will collect
   multiple copies of the same file. Note that some files naturally
   change between VSS copies (e.g. log files) so this can end up
   collecting a lot more data than anticipated.

   NOTE: Setting this will result in a slow down as we need to switch
   to using the `ntfs` accessor for all files (i.e. parse the low
   level filesystem), and inspect each VSS copy for a change in the
   file.

4. **MaxFileSize**: Sometimes we encounter very large files in
   unexpected location (e.g. browser cache). This setting ensures that
   very large files will not be collected. By default the setting is
   disabled (i.e. we collect any file size), but it is a good idea to
   limit it as very large files are not often useful.

5. **UPLOAD_IS_RESUMABLE**: This setting controls how uploads are send
   from the Velociraptor client to the server. When enabled, the
   client will send upload information in advance so that if the
   collection times out or the client is restarted, the uploads may be
   resumed.

   The setting only has an effect when collecting this artifact
   remotely from a client (i.e. does nothing for offline collections).

Following these parameters, there are many checkboxes for each
possible collection target.

The most useful `meta-targets` are the `SANS_Triage`, `KapeTriage`

## Artifact

<div style="max-height: 500px; overflow-y: auto; ">
<pre >
<code style="margin-top: -40px;font-size: medium;" class="language-yaml">
{{< insert "./Windows.KapeFiles.Targets.yaml" >}}
</code>
</pre>
</div>
