This document is a copy of the original proposal available in [google docs](https://docs.google.com/document/d/12yPyyJsaMEoTXDJ-S5dhPa24sFiBpnOo0QMoLSroA4s/edit#)

# Proposal
Have the mutating webhook exist as a separate component and its lifecycle be
managed via helmcharts.

# The current plan
- elastic-agent or apm-server install the webhook config
- the apm-server serves the webhook endpoint
- the apm-server uninstalls the webhook if it is removed

Discussion can be found here:
- [APM agent attachment on k8s](https://docs.google.com/document/d/1RiF56EZLHOB7Yoo_y-pnZHZQM-RY1xfhh2Cp4y_Qiv0/edit#)
- https://github.com/elastic/obs-dc-team/issues/566

##Argument for this plan
Initial setup via elastic-agent would be easier. The user already has installed
the elastic-agent and apm-server, why make them use something else?

# Arguments against (pro helm)
##It’s significantly more complex
While initial setup via elastic-agent would be the easiest, that is not
considering the full lifecycle of the webhook component.

Updating, removing, and configuring the webhook would all be easier if it were
managed via a helmchart and its own independent docker container.

The current proposal to install the webhook configuration requires leader
election between elastic-agents, and its removal is proposed to be triggered by
removal of the apm integration. Leader election sounds like a solved problem,
but is needlessly complicated for adding a resource to a kubernetes cluster.

Removal still needs to be solved. Differentiating between a
shutdown/removal/upgrading the server will be difficult and present challenges
that can be avoided entirely.

##Adding an extra layer between the users and kubernetes that will be hard to debug or fix
Helm is the de facto way for handling software on kubernetes clusters. The
additional work to introduce error handling and remediation in our software
will place an unnecessary burden on developers, both during the initial
development and while supporting customers when something inevitably goes
wrong.

##Not all k8s clusters are the same
User clusters are going to vary widely; offering a helmchart that they can
easily access and modify to suit their needs will be infinitely more helpful
than the alternative: the user copies the configuration from within the
apm-server/elastic-agent, modifies it, and manually applies the new config to
their k8s server.

##elastic-agent installation is already manual
The initial installation of elastic-agent is a manual process requiring
interacting directly with the kubernetes cluster:
https://www.elastic.co/guide/en/fleet/master/running-on-kubernetes-managed-by-fleet.html

The cluster operators are already interacting with kubernetes directly and,
most likely, using helm.

Also, sidebar: managing elastic-agent within a kubernetes cluster should
probably be distributed as a helmchart.

EDIT: apparently elastic-agent, fleet, and kibana are having helmcharts
developed

#Arguments against (pro separate component)

##The webhook serves a different purpose from the apm-server
From the discussion in a different issue:

APM Server (the instance on the node that's the current leader) would be
responsible for ensuring a Service and a MutatingWebhookConfiguration object is
created

https://github.com/elastic/obs-dc-team/issues/566#issuecomment-1047536236

This is not, and should not, be the responsibility of the apm-server. The
apm-server should manage ingesting events and indexing them into elasticsearch;
installing objects into a kubernetes cluster and handling customer app
configuration is not something it should be doing.

Configuring kubernetes and mutating pods on pod creation are entirely different
responsibilities. I see this idea as the most direct solution to getting a
webhook managed via elastic-agent, which (again) I think is inappropriate.
Cluster-level configuration should exist elsewhere.

Webhook is very simple and should always be available. apm-server is more
complicated and more likely to crash/mis-behave; there’s no reason that the
webhook should share its fate with apm-server. Webhook can be 3 simple pods in
a deployment. Low resource-use, redundancy, and can be scaled as necessary by
operators.

##The apm-server and webhook have different lifecycles
The apm-server is a slow moving target and any changes necessary for the
webhook are going to be slow moving and painful for users.

The apm-server and webhook have different deploy requirements and will change
on different intervals; locking them to one another is going to either be
disruptive to the apm-server, or prevent updates to the webhook. A change to
the webhook should not require a user to upgrade their apm-server; nor should
an upgrade to the apm-server uninstall and re-install the webhook config.
Additionally, the agents have a different release cycle from the stack, which I
think further supports why the webhook should be independent.

