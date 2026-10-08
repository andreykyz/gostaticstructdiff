# Changelog

All notable changes to `gostaticstructdiff` are documented in this file.

## [0.1.11]

### Fixed

- Cross-package wrapped primitives (e.g. `type NodeID = string` in an imported
  package) are now resolved correctly regardless of the directory the tool is
  invoked from.

  Previously `resolveImportTypeDefs` ran `go list -find -json <import>` in the
  tool's current working directory. When the tool was run outside the module
  that contains the input file (e.g. from the `gostaticstructdiff` repo against
  files of another project), `go list` failed with
  `no required module provides package`, so no imported type could be resolved.
  Unresolved qualified field types (e.g. `config.NodeID`) then fell back to the
  "assume it's a struct" classification in `types.Classify` and were rendered
  through the nested-struct path, generating references to non-existent diff
  types such as `config.NodeIDDiff` / `config.NodeIDPatch` /
  `config.ApplyNodeIDDiff`.

  `go list` is now executed in the directory of the parsed input file
  (`StructInfo.SourceDir`), so imports resolve against the input file's module
  and wrapped primitives are classified as basic comparable types, rendered
  as-is (consistent with how map keys of the same types are rendered).
