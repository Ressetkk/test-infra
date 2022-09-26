# Concept

Buildog is more mature of [kaniko-build](../development/kaniko-build). Key concepts of it are:
* scalable
* every build is made as completely isolated workspace
* it's easy to extend
* administrators can enforce some requirements upon images
  * default repository url(s)
  * default tag
* User does not have access to service account credentials
* user can't access or edit build jobs
* Support for image signing
* Support for image caching
* Supports building multiple variants of the same image
* Easy configuration using configuration file

Software consists of 4 basic components (for now):
* *controller* - takes care of scheduling build jobs and managing their lifecycle
* *signer* - takes care of signing images by supported back-ends based on the requests from controller
* *sidecar* - works along in the job pod and manages the log collection and sends them to external storage when the job is done
* *init* - init container for job pods that takes care of preparing build workspace

Those 4 components should work at basic stage for POC stage.

Schedulers are piece of code that take care of creating a job using a desired image. It can be custom image, or official one.
Schedulers can have different configurations based on the needs. Those should be modified inside the buildog configuration file.
Schedulers are only meant to take care of preparing a job. 

POC Phase:
* *kaniko-scheduler* - a scheduler that performs setting up of the k8s Job using official `kaniko/executor` image
  * Supports all functionalities of current kaniko-build

To build images user can only communicate with the `controller`, that will execute a build using a desired scheduler.
Only authenticated user can perform requests to the build. Initial stage - token-based authorization - only users with token can do requests.
Controller uses its own cluster authentication to schedule jobs based on the received request. If needed, it parses the variants.yaml file, asks schedulers to prepare jobs for those variants and creates them in isolated namespace.
Init container prepares downloads required code to build.
Sidecar fetches logs from the build container, and once the build finishes, it downloads all logs and pushes it to the external storage.
Controller fetches an information about finished job and notifies the user that the build is done.

Things to consider:
* How should we send job request to the controller?
* Should we provide CLI that allows running builds from shell?
* Should we provide GitHub webhook implementation for file-based configurations?

```mermaid
C4Context
title Buildog implementation

```