##Changelog

### Flotilla 2.0.0 (20.1.2016)

- new internal package structure
- refactored App Environment & Configuration
- rewrite and separate packaging into more distributed libraries for App,
  Assets, Extensions, Logging, Blueprints, Routes, Templates, Flash, etc.
- changed 'Ctx' to 'State', including refactoring and a reversion to a fuller
  interface with less emphasis on extension functions(but still retaining the
  extension functionality for relevant items)


### Flotilla 1.1.0 (24.6.2015)

- refactored testing
- Expectation & Performer based testing idioms
- improvements to Blueprint registry
- improvements to Route creation & management
- updated & refactored Flash messaging


### Flotilla 1.0.1 (19.3.2015)

- bugfix for inability to add multiple Manage for different request methods, but the same path


### Flotilla 1.0.0 (18.3.2015)

- [ground-up reassembly & rewrite from former codebase](https://github.com/fc-thrisp-hurrata-dlm-graveyard/flotilla-defunct)
- eliminate dependency on former [Engine](https://github.com/fc-thrisp-hurrata-dlm-graveyard/engine-defunct) package
- routing subpackage: engine defining Engine interface, for per app router configuration
- Ctx interface, supporting per Blueprint ctx configuration
- revised Ctx to support [context.Context](https://github.com/golang/net/tree/master/context) interface (proof of concept requiring more work)
- new extension interface for extension: Fxtensions(`Flotilla Extension`) to manage & test multiple ctx extension functions as a unit.
- improved testing
- errors subpackage: xrr defining an interface for uniform, reusable erroring 
- refactor, cleanup and subsequent new messes overall 


### Flotilla 0.3.2 ~unreleased

- change 'ctx functions' to 'extensions'


### Flotilla 0.3.1 (16.12.2014)

- update travis.yml to accomodate 1.4/1.3 cover package path difference


### Flotilla 0.3.0 (15.12.2014)

- new Blueprint concepts 
- eliminate old Blueprint interface, merge RouteGroup & Blueprint idioms to one
- engine interface & default engine, for future extensible engines
- essential testing, bugfixes, and refactoring


### Flotilla 0.2.0 ~unreleased

- return from 'R' to 'Ctx'
- flash messaging
- per-route, in-template context processors
- initialization & configuration streamlining
- coinciding boolean app modes (development, testing, & production)
- methods for viewing & setting plain or minimally secure cookies
- public store item value for easier access to store settings 
- deferral of ctx functions until after all handlers have run
- bugfixes & refactoring  


### Flotilla 0.1.0 (17.10.2014)

- basic adherence to semantic versioning
- 'R' type for per-route handled context
- reintegrate Djinn(formerly Jingo) templating
- tighter interaction with Engine statuses & panics
- simple configuration functions, removal of flags 
- package-level errors
- essential testing, bugfixes, and refactoring  


### Flotilla 0.0.2 (24.9.2014)

- extend Ctx with cross handler functions
- simple flag parsing for run mode (production, development, testing)
- cookie based sessions as a default with capacity for adding different backends
- folded router & some lower-level-but-not-in-net/http(such as http statuses)
  functions into another package: [Engine](https://github.com/thrisp/engine)
- url formatting by route, i.e. creating urls by route name & parameters


### Flotilla 0.0.1 (20.8.2014)

- reforked, renamed as Flotilla
- ini style configuration read into app environment
- basic Flotilla interface for extension of routes & env
- basic Jingo templating
- provisions for binary static or template Asset inclusion per engine
 

### Fleet 0.0.0 (22.7.2014)

- forked from https://github.com/gin-gonic/gin
