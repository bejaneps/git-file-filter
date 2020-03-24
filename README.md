# Git File Filter

A webapp that filters git repository files using regexp.

## Building requirements

### Windows
----------------------------------------------------------

* [Go](https://golang.org/) -  main dependency of builder

### Linux
----------------------------------------------------------
* [Go](https://golang.org/) -  main dependency of builder

## Compiling and running

1. Check if you have go installed on your machine by typing:

    ```
    $ go version
    ```

2. Run a *build.sh* script:
    ```
    $ ./build.sh
    ```

3. If you don't prefer running scripts or you are using Windows, then run an app directly:
    ```
    $ go run ./cmd/web/main.go
    ```

4. If a user wants to run a server on different port, he can do it by changing _PORT_ environment variable. To change it type:
    ```
    $ export $PORT=1234
    ```

## Description

There are 4 pages in total, each page has it's own function. Main entrance is a **Search** page, where user first have to fill the form and send it to server, after that server parses all repository structure and saves it in cache for later use. For filtering specific files, e.g: config files, one can specify a regexp pattern in **Filter** page (in .json format) and then submit the pattern to server, result is saved in cache and can be seen by user in **Configs** page.

**Search** - user types an absolute url of git repository and all the files in that repository are shown in _Files_ page. Commit hash and Directory are _optional_, if user didn't fill commit hash field, server will use latest commit(head). If user didn't fill directory field, server will use root directory.

**Filter** - user types a regexp pattern in json form, and the server filters files in a repository (or in a specific folder) and puts them in **Configs** page. Additionally it offers user to download a result file in json format. Example:

    
    {
        "config": [
            {
                "name": "Docker",
                "filter": "\b(docker-compose.yml|docker-compose.yaml)\b",
                "policy": "https://example.com/docker.rego"
            },
            {
                "name": "Terraform",
                "filter": "\b(.tf|.tf.json)\b",
                "policy": "https://example.com/terraform.rego"
            }
        ]
    }
    

**NOTE:** policy field is optional, if it's not mentioned, then an app will try to search a policy in git repo, if it doesn't find it, then it will user default policy.

**Files** - all the files in a root or specific directory of a repository are shown here. Each file name has a link to it's git location, as well as it's hash.

**Configs** - all the files that were filtered by regexp are shown here. In addition to file names, content of files are also shown here.

**NOTE:** before making a new filter request, user should search a repository in **Search** page, otherwise he will be redirected to search page.