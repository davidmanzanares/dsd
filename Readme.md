# Dead Simple Deploy
`dsd` is a simple tool to deploy assets (binary programs, scripts, data) to remote computers.

## Configuring new deployments
```
$ dsd init
dsd configured at /home/david/myCode/myAwesomeProject/

$ dsd add dev "s3://myAwesomeBucket/dev/" "myBinary" "*/*.glsl" "*/*.txt" "*/*.ttf" "*/*.ogg"
Target "dev" (s3://myAwesomeBucket/dev/) {"myBinary", "*/*.glsl", "*/*.txt", "*/*.ttf", "*/*.ogg"} added
```

## Deploying
```
$ dsd deploy dev
Deploying to "dev" (s3://myAwesomeBucket/dev/) {"myBinary", "*/*.glsl", "*/*.txt", "*/*.ttf", "*/*.ogg"}
Deployed  {2020-03-07T00:13:52Z #e89c69676dfe0659 2020-03-07 01:13:53.536911707 +0100 CET m=+1.529182466}
```

## Getting updates (deploys)
```
$ dsd watch "s3://mydeploybucket/dev"
Started up version [eb08c683f2c4fba93b31afaac77f9bc556e2a0bf]
```
