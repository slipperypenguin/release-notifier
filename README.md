# release-tracker
[WIP]

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
go build
./release-tracker
```

## Deployment
**Kubernetes + Helm**

Create a secret with the required credentials:
```shell script
kubectl create secret generic release-tracker \
        --from-literal=github=XXX` \
        --from-literal=slack=XXX
```
You can then install the deployment through Helm:
```shell script
cd release-tracker
helm upgrade -i release-tracker
```


### Credits
Many thanks to [justwatchcom](https://github.com/justwatchcom)!
