# CONEX

Website aggregator

## Roadmap

### Sites manger [4/4]

- [x] Define API interface for creating, (un)publishing and deleting sites.
      I'm guessing, uploading HTML (for now don't sanitize, assume ok) and a path
      for the site to be published. A companion request in this folder for a GET
      request that checks if site path is available would be nice, maybe
      GET `/sites/available/{path}`, and error if it exists (or is used by
      other endpoints), 204 No Content if it does not already exist and is
      available.
      In order to check if it exists I think it's convenient to use the
      in-memory objects from storage.Storage interface, since it implements
      Exists(key string) bool.
      The endpoint for (un)publishing could just be a PATCH (of course this one
      requires access token) with `{"public":true/false}`.
      Notice `/sites/` and other handlers would then be in h.paths, use that
      plus checking with the other method if a site already exists to check if the
      path is available.
      When uploading a site, upload with content-type text/html to the path
      AFTER successfully registering in the db that path as owned by the user.
- [x] Update `database/sqlite` and `database/postgres` submodules to allow users
      to create sites, a site would consist of a path VARCHAR(255) and public
      boolean status. The path would also locate the site in a s3 bucket.
      For example, the path could be `go-fitness`, and so, if present in a
      bucket, it's key would be `go-fitness`.
- [x] Create `sites/site.go` submodule, with a `Sites` interface that abstracts
      the logic required to create, (un)publish and delete sites, this submodule
      would wrap the database and object storage logic.
- [x] Implement handlers that meet the API interface for sites.
