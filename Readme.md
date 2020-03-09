# Dead Simple Deploy
`dsd` is a simple tool to deploy assets (binary programs, scripts, data...) to remote computers.

## Installing

- Install Go: https://golang.org/dl/
- Clone me: `git clone git@github.com:davidmanzanares/dsd.git`
- cd to the CLI directory: `cd dsd/dsd`
- Compile me: `go build`
- The `dsd` program should have been created

### Installing it on your path
- Add Go's bin folder to your path: `export PATH=$PATH:$(go env GOPATH)/bin` (see https://golang.org/doc/gopath_code.html)
- Use `go install` in you `dsd/dsd` folder

## Configuring new deployments
```

$ dsd add dev "s3://myAwesomeBucket/dev/" "myBinary" "*/*.glsl" "*/*.txt" "*/*.ttf" "*/*.ogg"
Target "dev" (s3://myAwesomeBucket/dev/) {"myBinary", "*/*.glsl", "*/*.txt", "*/*.ttf", "*/*.ogg"} added
```

## Deploying
```
$ dsd deploy dev
Deploying to "dev" (s3://myAwesomeBucket/dev/) {"myBinary", "*/*.glsl", "*/*.txt", "*/*.ttf", "*/*.ogg"}
Deployed  {2020-03-07T00:13:52Z #e89c69676dfe0659 2020-03-07 01:13:53.536911707 +0100 CET m=+1.529182466}
```

## Running the deployed packages

Run once:
```
$ dsd run "s3://mydeploybucket/dev"
AppStarted{v: {2020-03-08T15:36:54Z #46dcf80b9c7cbbd8 2020-03-08 16:36:55.43163728 +0100 CET}}
```

Run and restart every time the application exits:
```
$ dsd run --on-success restart --on-failure restart "s3://mydeploybucket/dev"
AppStarted{v: {2020-03-08T15:36:54Z #46dcf80b9c7cbbd8 2020-03-08 16:36:55.43163728 +0100 CET}}
```

Run and restart every time the application exits, restart the application with new updates:
```
$ dsd run --on-success restart --on-failure restart --hotreload "s3://mydeploybucket/dev"
AppStarted{v: {2020-03-08T15:36:54Z #46dcf80b9c7cbbd8 2020-03-08 16:36:55.43163728 +0100 CET}}
```

Run and start the application again with new updates when the app exits:
```
$ dsd run --on-success wait --on-failure wait "s3://mydeploybucket/dev"
AppStarted{v: {2020-03-08T15:36:54Z #46dcf80b9c7cbbd8 2020-03-08 16:36:55.43163728 +0100 CET}}
```
