# T00–T27 archival branch

This branch contains the five independently hash-checked source packages for T00–T27 and the T27-R1 external sidecar. The archive files are tracked with Git LFS. GitHub rejected the 2,417,333,307-byte T00–T23 ZIP because its per-file LFS limit is 2,147,483,648 bytes. Therefore that original file is preserved as ordered `.part.00` and `.part.01` LFS fragments; concatenating them in lexical order reconstructs the exact original package SHA-256 listed in `T27_T00_T27_PACKAGE_MANIFEST.sha256` and `T00_T27_PACKAGE_CHUNKS.sha256`.

No business code was changed; no runtime, Camoufox, real proxy, real cookie, native SystemVault, release, or gate operation was performed by this archival addition.
