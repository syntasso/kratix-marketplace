# Waypoint Promise

This Promise provides [Waypoint](https://www.waypointproject.io/) access as-a-Service. The Promise installs a Waypoint server which then can vend invite tokens through Resource Requests.

The Promise has 1 field:
- `.spec.username`: This is the username that will be provided an invite token for access to Waypoint

Check the CRD documentation for more information.

To install:

> **Warning**
> 
> **By default Waypoint requires a LoadBalancer Service type**
> 
> **If you are running [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/) or any other cluster without support by default, please either:**
> **1. Follow cluster software instructions to install LoadBalancers (e.g. [here](https://kind.sigs.k8s.io/docs/user/loadbalancer/))**
> **2. Change this Promise to use NodePort type Services (see: [Developer Readme](./internal/README.md#switch-to-nodeport))**


```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/waypoint/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/waypoint/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
