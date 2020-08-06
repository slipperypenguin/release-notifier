Instructions for connecting to Private GitHub Reigstry (beta access)


1. Create new Github Personal Access Token with `read:packages` scope.

2. Base-64 encode `<your-github-username>:<TOKEN>`,
  ie.:
  ```
  $ echo -n slipperypenguin:6111aaab222ab333aa4447 | base64
  <AUTH>
  ```

3. Create .yaml file that can be used in `kubectl apply -f`:

4. Reference the above secret from your pod's spec definition via `imagePullSecrets` field
