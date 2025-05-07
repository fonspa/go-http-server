# Docs

## Monolith vs Decoupled Architecture

Monolith is a single large program that contains all the functionality for both front-end and back-end.

![monolith architecture](./monolith.png)

Decoupled architecture is one where front-end and back-end are separated in different codebases.

![decoupled architecture](./decoupled.png)

Pros and Cons:
- Monolith
  - Simpler
  - Easier to deploy new version when everything in the app is always in sync
  - Data being embedded in the HTML can lead to better UX
- Decoupled
  - Easier to scale
  - Better separation of concern as codebase grows
  - Siter and API can be hosted on separate servers using separate technologies
  - Embedding data in HTML is possible with pre-rendering but more complicated

=> When building a new app from scratch, start with a monolith, but keep front-end and API decoupled logically within the project from the start.
