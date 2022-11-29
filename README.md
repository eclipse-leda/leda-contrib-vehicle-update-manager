# Vehicle Update Manager
This repository contains the vehicle update manager component used in the Software Defined Vehicle EDGE stack.

The vehicle update manager (VUM) is a component responsible for the orchestration of OTA Updates towards the vehicle. It is an extended and adapted version of the [Eclipse Kanto's Container Manager](https://github.com/eclipse-kanto/container-management) that is being able to handle new desired state for the software on the whole vehicle. 

The desired state comes in a descriptive way as a multi document YAML content, a.k.a desired state manifest, and it includes a list of Kubernetes resources (Deployments, Pods, Services, ConfigMaps, custom resources, etc.). 

VUM detects the system-level update custom resource and passes it for further processing to the [Self Update Agent](https://github.com/eclipse-leda/leda-contrib-self-update-agent). The remaining resources are forwarded to a Kubernetes control plane and handled like the well-known _kubectl_ command - creating new resources, updating existing ones or deleting old ones that are no longer present in the desired state manifest. VUM also monitors the self-update agent and the control plane, and compiles and report the current state of the device, again as a list of Kubernetes resources.
# Build
## Build prerequisites
The following libraries need to be available on the build host:
- GNU Make 
- Go 1.17.2 (or higher)
## Performing a build
Navigate to the `updatem` directory and call (no super user privileges required) :
```commandline
make clean && make build-targets-archive
```
# Install
## Runtime prerequisites
- Kubernetes distribution, e.g. k3s
## Configuration prerequisites
As the Vehicle update manager is an extended and adapted version of the [Eclipse Kanto's Container Manager](https://github.com/eclipse-kanto/container-management), it is represented to the outside world as a Ditto Thing and it needs the proper configuration of its Ditto Features. 

The following configuration must be set in **/etc/updatemanagerd/updatemanagerd-config.json**:
```json
{
   "things": {
        "features" : ["UpdateOrchestrator"]
   }
}
```
## Performing an installation
After a successful build has been performed, from the same `updatem` directory call:
```commandline
sudo make install
```
# Run
Start the updatemanagerd daemon:
```commandline
sudo systemctl daemon-reload
sudo systemctl enable updatemanagerd.service
sudo systemctl start updatemanagerd.service
```

# Containerized version
A containerized version of the VUM binary can be built and run as well.
The Eclipse Leda distribution makes use of VUM as a container running through a Kuberneter deployment.
# Contributing

If you want to contribute bug reports or feature requests, please use *GitHub Issues*.
For details, please see our [contributing guidelines](CONTRIBUTING.md)

# Data privacy notice
Please see our [data privacy notice](DATA_PRIVACY_NOTICE.md)

# License and Copyright

This program and the accompanying materials are made available under the
terms of the Apache License 2.0 which is available at
https://www.apache.org/licenses/LICENSE-2.0

For details, please see our license [LICENSE](LICENSE)
