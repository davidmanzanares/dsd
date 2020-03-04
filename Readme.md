# Dead Simple Deploy
`dsd` is a simple tool to deploy assets (binary programs, scripts, data) to remote computers.

## Configuring new deployments
```
$ dsd init
dsd configured at /home/david/myCode/myAwesomeProject/

$ dsd add dev "s3://mydeploybucket/dev" "myAwesomeBinary"
Added "dev" (s3://mydeploybucket/dev) {"myAwesomeBinary"}
```

## Deploying
```
$ dsd dev
Deployed [eb08c683f2c4fba93b31afaac77f9bc556e2a0bf] to "dev" (s3://mydeploybucket/dev) {"myAwesomeBinary"}
```

## Getting updates (deploys)
```
$ dsd watch "s3://mydeploybucket/dev"
Started up version [eb08c683f2c4fba93b31afaac77f9bc556e2a0bf]
```

## Advanced configuration

### Controlling the start/stop cycle

```
$ dsd conf lifecycle.relaunch true dev
```
