---
title: "Linux.Triage.UAC"
date: 2025-08-11
weight: 20
bookToc: false
IconClass: fa-solid fa-desktop
---

# Linux.Triage.UAC

This artifact is built automatically from the [UAC
project](https://github.com/tclahr/uac) project.

You can [download the artifact](Linux.Triage.UAC.zip) for manual
import into Velociraptor.

The description below explains how to use this artifact in practice.

The artifact will generate a list of globs and prepend the device name
to each glob. Velociraptor's `glob()` plugin implementation is very
efficient and minimizes the number of passes it needs to make over the
filesystem, when using multiple glob expressions at the same time.

Therefore the artifact first traverses all the rules to build a large
list of glob expressions, which it uses to search for candidate files.

## Parameters

1. **MaxFileSize**: Sometimes we encounter very large files in
   unexpected location (e.g. browser cache). This setting ensures that
   very large files will not be collected. By default the setting is
   disabled (i.e. we collect any file size), but it is a good idea to
   limit it as very large files are not often useful.

2. **UPLOAD_IS_RESUMABLE**: This setting controls how uploads are send
   from the Velociraptor client to the server. When enabled, the
   client will send upload information in advance so that if the
   collection times out or the client is restarted, the uploads may be
   resumed.

   The setting only has an effect when collecting this artifact
   remotely from a client (i.e. does nothing for offline collections).

Following these parameters, there are many checkboxes for each
possible collection target.

## Artifact

<div style="max-height: 500px; overflow-y: auto; ">
<pre >
<code style="margin-top: -40px;font-size: medium;" class="language-yaml">
{{< insert "./Linux.Triage.UAC.yaml" >}}
</code>
</pre>
</div>
