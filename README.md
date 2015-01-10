# goinx

Just a 80-lines domain proxy mapping service.

Update the mapping by written an router.conf(default, modified by -c)

```
google.com localhost:8080
github.com localhost:8900
```

Each rule a line and seperated by a space.

Visit /ref@ to notify the server to reread the router file.

```
curl localhost/ref@
```
