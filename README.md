# CONEX

## To Do
  - [X] update templates like `templates.LoginWarnInvalidEmail` to `templates.LoginWarn` and pass the required messages via translator
  - [X] use middleware to easily scope authentication, for example `handlers.verifyClient` could be middleware
  - [X] user plans table
  - [X] Send CSRF token in client
  - [X] Account logins, devices, etc
  - [X] site registration
  - [X] site update
  - [X] Finish templates conditions
  - [X] Lock down endpoints
  - [X] Tags management
  - [X] S3 / Upload images
  - [X] site sync
  - [X] Clean logger, add useful info
  - [X] New/Update banner
  - [X] payments
  - [X] allow toggling home page
  - [X] check image endpoints are actually images
  - [X] check for plans in endpoints
  - [X] terms
  - [X] check for site quota
  - [X] Loading screen for editor on slow connections
  - [X] ! allow email change
  - [X] ! allow account delete
  - [X] ! allow site delete
  - [ ] ! search
  - [ ] ! Harden against db limits
  - [ ] spinner on subscribe
  - [ ] filter wordlists
  - [ ] Auto-disable plans
  - [ ] check for image quota per site
  - [ ] themes

## bugs
  - [ ] tags on first load with no tags when closing the modal still does $showtags until reload with tags
