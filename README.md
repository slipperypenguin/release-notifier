# release-notifier
🛰 slack notifications for OSS releases

Receive slack notifications for specified GitHub releases. Better than the vanilla RSS notifier


## Getting setup
**Credentials**
1. Slack [incoming webook](https://api.slack.com/incoming-webhooks) url
2. GitHub API Token ([docs](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line))

**Watching repositories**
- Track repositories by adding them to the list of armuments:
  - `-r=golang/go -r=kubernetes/kubernetes` etc...


**Running Locally**
```go
go build -mod=vendor
./release-notifier
```

## Deployment
### **Building + pushing locally**
General
```
$ docker build -t docker.pkg.github.com/slipperypenguin/release-notifier/release-notifier:v0.1.X .
$ docker images
$ docker push docker.pkg.github.com/slipperypenguin/release-notifier/release-notifier:v0.1.X
```

Example ARM
```
$ docker build -f DockerfileARM -t docker.pkg.github.com/slipperypenguin/release-notifier/release-notifier:v0.1.2 .
$ docker images
$ docker push docker.pkg.github.com/slipperypenguin/release-notifier/release-notifier:v0.1.2
```


### **Kubernetes + Helm**
Populate the `values.yaml` file with the required `slack` and `github` credentials. This normally should be handled by a CI tool such as Jenkins or Travis. If these keys are manually added to `values.yaml`, make sure they are not committed.

You can then install the deployment through Helm.

Create a deployment + secret with the required credentials:
```shell script
cd deployment
helm upgrade -i release-notifier ./release-notifier --dry-run
```


### Credits
Many thanks to [justwatchcom](https://github.com/justwatchcom)!
