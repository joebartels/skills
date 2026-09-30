# Catalog export library

Review the complete code area, including existing in-scope issues, across catalog and export packages. This is a library, with no deployment pipeline, service, external I/O or third-party dependencies. Go 1.21 is the supported language version. The public Catalog.Snapshot contract promises independently mutable results; caller modifications must not modify the catalog. Export.Uppercase relies on that contract and must not mutate stored labels. Review both packages and their interaction. Tests and all relevant source are supplied here.
