---
title: Plugins as services
authors:
  - "@Joibel" # Authors' github accounts here.
sponsors:
  - "@crenshawdev"?
  - TBD        # List all interested parties here.
reviewers:
  - "@alexmt"
  - TBD
approvers:
  - "@alexmt"
  - TBD

creation-date: yyyy-mm-dd
last-updated: yyyy-mm-dd
---

# Plugins as Services

Allow Argo CD to discover and use plugins via kubernetes services, in the same way as sidecars.

<!-- ## Open Questions [optional] -->

<!-- This is where to call out areas of the design that require closure before deciding to implement the -->
<!-- design. -->

## Summary

As a complement to sidecar deployment of plugins, allow the repo-server to discover services in the same namespace as it with a label `argocd.argoproj.io/plugin: true` and communicate with them.

Plugins already communicate via gRPC and would continue to use the same protocol for all purposes, this proposal does not change the protocol nor capabilities.

## Motivation

This would allow plugins to be deployed independently of Argo CD as separate applications, with separate lifecycles. Plugins could chose their own preferred deployment mechanism, such as helm charts, and could come with assisting deployments of HPAs.

### Goals

It would be possible to deploy a plugin as a minimal set of service+deployment (using an existing plugin) and the use of that plugin would be indistinguishable from the deployment as a sidecar.

An example of how a plugin might be deployed this way would be documented.

### Non-Goals

Implementation of any actual plugins except examples necessary to prove the mechanism is working. In reality, as a plugin author, I'd implement it in [argocd-lovely-plugin](https://github.com/crumbhole/argocd-lovely-plugin/).

Implementation of any changes to the protocol or capabilities of plugins.

## Proposal

The repo-servers would discover plugins by Listing and Watching services with a specific label within the deployment namespace. Upon discovery the plugin would be queried over gRPC for the metadata, in the same way as a sidecar - and the process would proceed from there.

Deletion of a plugin would need to be detected and handled within the repo-server. This currently cannot happen.

### Use cases

As a plugin author I would like to make it easier for users to install my plugin, and remove it if they don't like it - trialling a plugin would become easier and reduce the apparent risk of modifying my argocd-repo-server.

Plugins would be scalable independently of repo-servers. It would be trivial to make plugins horizontally scale with an HPA, in a way that is not currently easy to do with the repo-server. (Future possibility of officially offloading helm and kustomize to HPA controlled deployments may aid scalability - this is not part of this proposal).

Plugin unavailability would not affect my entire repo-server (at the moment if one of your plugin images becomes unavailable due to the repository being down, the whole repo-server pod won't start).

As a plugin developer I'd like to be able to test and iterate faster - I can now do this with my own deployment or use of an `ExternalName` service pointing at my personal development environment.

Plugins could also have their own `ServiceAccount` and appropriate naughty roles to be able to do horrid things using the cluster as a source of data without having to grant those permissions to the repo-server. Perhaps a nice example would be to allow use of secrets.

### Implementation Details/Notes/Constraints [optional]

The repo-server<>plugin traffic would be between pods, and there is no built in mechanism for this to be secured. This is no worse than the default traffic between the rest of the Argo CD's components, and can be mitigated by deploying atop a service mesh if you don't trust your cluster.

Perhaps we should consider allowing configuration of which namespaces to look in, with multiple namespaces possible.

### Detailed examples



### Security Considerations

* There is an increased risk of interception and manipulation of the gRPC calls between the repo-server and plugin. I do not propose to mitigate this, as the same concern already exists between other components with a similar level of potential damage causable.

### Risks and Mitigations

I have nothing to add here.


### Upgrade / Downgrade Strategy

This is only a new mechanism, downgrade would prevent use of the new mechanism.

## Drawbacks

As with all features, it adds complexity to the product.

## Alternatives

This is already an alternative to an existing mechanism. I don't have an alternative proposal.
