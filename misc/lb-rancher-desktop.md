In Rancher Desktop Kubernetes, when you create a service of type `LoadBalancer`, it can sometimes get stuck in the "Pending" state for `External-IP` because Rancher Desktop doesn't have a built-in cloud provider to provision a load balancer like managed cloud services do (e.g., AWS or GCP). Here are a few alternatives to assign an `External-IP` to your service:

### Option 1: Use a `NodePort` Service
The simplest option is to expose your service using `NodePort`. This allows you to access the service through the node’s IP address on a high-numbered port.

1. Update your service type to `NodePort` in your YAML:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: golang-app-service
spec:
  selector:
    app: golang-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
      nodePort: 30007  # Specify a port in the range 30000-32767 (or leave empty to auto-assign)
  type: NodePort
```

2. Apply the service:

   ```bash
   kubectl apply -f app-service.yaml
   ```

3. Find the IP address of your node:

   ```bash
   kubectl get nodes -o wide
   ```

4. Access your service via `http://<node-ip>:<nodePort>`. For example:

   ```bash
   http://localhost:30007
   ```

### Option 2: Use `kubectl port-forward`
You can also use `kubectl port-forward` to expose a specific service or pod to your local machine. This doesn't assign an `External-IP`, but it's useful for local development.

1. Forward the service port to your localhost:

   ```bash
   kubectl port-forward svc/golang-app-service 8080:80
   ```

2. Access the service at `http://localhost:8080`.

### Option 3: Use `MetalLB` for LoadBalancer Support
If you want to assign an actual `External-IP` using a `LoadBalancer` service, you can install **MetalLB**, a load balancer implementation for bare-metal Kubernetes clusters. It works with Rancher Desktop by providing IP addresses for `LoadBalancer` services.

#### Steps to Install MetalLB:

1. **Install MetalLB**:

   ```bash
   kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.10/config/manifests/metallb-native.yaml
   ```

2. **Configure an IP Address Pool**:

   You'll need to define an IP range that MetalLB can assign as `External-IP`s. Create a `ConfigMap` for MetalLB:

   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     namespace: metallb-system
     name: config
   data:
     config: |
       address-pools:
       - name: default
         protocol: layer2
         addresses:
         - 192.168.1.240-192.168.1.250  # Example range of IP addresses in your network
   ```

   Apply the `ConfigMap`:

   ```bash
   kubectl apply -f metallb-config.yaml
   ```

3. **Update Your Service Type to `LoadBalancer`**:

   Update your service to use `LoadBalancer` as the type:

   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: golang-app-service
   spec:
     selector:
       app: golang-app
     ports:
       - protocol: TCP
         port: 80
         targetPort: 8080
     type: LoadBalancer
   ```

4. **Deploy the Service**:

   ```bash
   kubectl apply -f app-service.yaml
   ```

   MetalLB will now assign an IP from the configured range as the `External-IP` for your service.
