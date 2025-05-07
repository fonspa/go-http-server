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
  - Site and API can be hosted on separate servers using separate technologies
  - Embedding data in HTML is possible with pre-rendering but more complicated

=> When building a new app from scratch, start with a monolith, but keep front-end and API decoupled logically within the project from the start.

## Deployment Options

### Monolith

Because it's just one program, you just need to get it running on a server exposed to the internet and point the DNS records to it.

Cloud platforms options:
- AWS EC2, GCP Compute Engine, Digital Ocean Droplets, Azure Virtual Machines, Heroku, Google App Engine, Fly.io, AWS Elastic Beanstalk...

### Decoupled

Here there are two different programs that need to be deployed. You could deploy both on the same kind of places, or use a specific platform for the front-end:
- Vercel, Netlify, Github pages...

## Testing with Curl

Post Json with Curl
```shell
curl -X POST -L -H "Content-Type:application/json" -d '{"key":"value"}' <url>
```
