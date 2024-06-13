This repository provides a simple module for implementing delegate tasks. This repository comes with a sample `main.go` executable that demonstrates how this module can be used.

Exec a script (built-in):

```sh
./go-task --pretty samples/sample-exec.json
```

Get a secret from a file (built-in):

```sh
./go-task --pretty samples/sample-file.json
```

Get a secret (built-in) and use in script (built-in):

```sh
./go-task --pretty samples/sample-file-echo.json
```

Get a multi-line secret (built-in) and use in script (built-in):

```sh
./go-task --pretty samples/sample-file-echo-multiline.json
```

Get a user (cgi):

```sh
./go-task --pretty samples/get-user/task.json
```

Get a user list (cgi):

```sh
./go-task --pretty samples/get-user-list/task.json
```

Get a secret from vault (cgi):

```sh
./go-task --pretty samples/get-secret/task.json
```

Get a secret from vault (built-in):

```sh
./go-task --pretty task/secret/vault/fetch.json
```

Get a user (cgi) using a remote artifact (git):

```sh
./go-task --pretty samples/sample-cgi-from-repo.json
```
